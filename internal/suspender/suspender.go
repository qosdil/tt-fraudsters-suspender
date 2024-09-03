package suspender

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"main/internal/cognito"
	"main/internal/database"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
)

func (s *Suspender) BatchSuspend(ctx context.Context, buf bytes.Buffer, status BatchSuspensionStatus) (BatchSuspensionStatus, error) {
	scanner := bufio.NewScanner(&buf)
	suspensionStatuses := make(chan SuspensionStatus, status.NumRecords)

	var wg sync.WaitGroup
	var userID string
	var err error

	rowsPerChunk, err := s.getRowsPerChunk()
	if err != nil {
		return status, err
	}

	for scanner.Scan() {
		wg.Add(1)
		userID = scanner.Text()

		// Wait for a second if number of chunked rows exceeds rowsPerChunk
		// This is to overcome RPS limitation on Cognito
		if numChunkRows >= rowsPerChunk {
			time.Sleep(time.Second)
		}

		numChunkRows++
		go s.ConSuspend(ctx, userID, suspensionStatuses, &wg)
	}

	go func() {
		wg.Wait()
		close(suspensionStatuses)
	}()

	for suspensionStatus := range suspensionStatuses {
		if suspensionStatus.Error == nil {
			status.NumSuccessful++
			continue
		}

		status.NumFailed++
		err = fmt.Errorf("failed suspending user %s: %s", suspensionStatus.UserID, suspensionStatus.Error.Error())
		status.Failures = append(status.Failures, err)
	}

	return status, nil
}

// ConSuspend suspends user data with concurrency
func (s *Suspender) ConSuspend(ctx context.Context, userID string, suspensionStatus chan SuspensionStatus, wg *sync.WaitGroup) (err error) {
	defer wg.Done()
	if err = s.Suspend(ctx, userID); err != nil {
		suspensionStatus <- SuspensionStatus{UserID: userID, Error: err}
		numChunkRows--
		return err
	}

	suspensionStatus <- SuspensionStatus{UserID: userID}
	numChunkRows--
	return nil
}

func (s *Suspender) getRowsPerChunk() (rowsPerChunk float32, err error) {
	cognitoMaxRPS64, err := strconv.ParseFloat(os.Getenv(envCognitoMaxRPS), 32)
	if err != nil {
		return 0, fmt.Errorf("failed to parse env var of %s: %s", envCognitoMaxRPS, err.Error())
	}
	cognitoMaxRPS := float32(cognitoMaxRPS64)

	cognitoMaxRPSChunkRatio64, err := strconv.ParseFloat(os.Getenv(envCognitoMaxRPSChunkRatio), 32)
	if err != nil {
		return 0, fmt.Errorf("failed to parse env var of %s: %s", envCognitoMaxRPSChunkRatio, err.Error())
	}
	cognitoMaxRPSChunkRatio := float32(cognitoMaxRPSChunkRatio64)

	return cognitoMaxRPS * cognitoMaxRPSChunkRatio, nil
}

func (s *Suspender) Suspend(ctx context.Context, userID string) (err error) {
	if err := s.Cognito.DisableUser(ctx, userID); err != nil {
		return err
	}
	if err := s.Database.DisableUser(userID); err != nil {
		// Re-enable user on Cognito
		s.Cognito.EnableUser(ctx, userID)

		return err
	}
	return nil
}

func (s *Suspender) CreateBufFromFile(ctx context.Context, sourceFile string) (batchBuffer BatchBuffer, err error) {
	// Open the source file
	file, err := os.Open(sourceFile)
	if err != nil {
		return batchBuffer, err
	}
	defer file.Close()

	// Scan the content
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	var userID string

	// Count number of lines, validate UUID v4
	line := 0
	for scanner.Scan() {
		userID = scanner.Text()
		line++
		if _, err := uuid.Parse(userID); err != nil {
			return batchBuffer, fmt.Errorf(`"%s" on line %d is not a valid UUID v4`, userID, line)
		}
		batchBuffer.NumRecords++

		// Collect user IDs
		fmt.Fprintln(&batchBuffer.Buf, userID)
	}

	return batchBuffer, nil
}

var numChunkRows float32

func NewSuspender(cognito *cognito.Cognito, sqlDB *database.Database) *Suspender {
	s := new(Suspender)
	s.Cognito = cognito
	s.Database = sqlDB
	return s
}

const (
	DoneMsg                    = "batch suspension done, # of records: %d, # of successful: %d, # of failed: %d\n"
	envCognitoMaxRPS           = "AMAZON_COGNITO_MAX_RPS"
	envCognitoMaxRPSChunkRatio = "AMAZON_COGNITO_MAX_RPS_CHUNK_RATIO"
)

type BatchBuffer struct {
	Buf        bytes.Buffer
	NumRecords int
}

type BatchSuspensionStatus struct {
	NumRecords    int
	NumSuccessful int
	NumFailed     int
	Failures      []error
}

type Suspender struct {
	Cognito  *cognito.Cognito
	Database *database.Database
}

type SuspensionStatus struct {
	UserID string `json:"user_id"`
	Error  error  `json:"error"`
}
