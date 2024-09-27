package suspender

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"sync"
	"time"
	"tt-fraudsters-suspender/internal/datastores/cognito"
	database "tt-fraudsters-suspender/internal/datastores/postgres"

	"github.com/google/uuid"
)

func (s *Suspender) BatchSuspend(ctx context.Context, buf bytes.Buffer, status BatchSuspensionStatus) (BatchSuspensionStatus, error) {
	scanner := bufio.NewScanner(&buf)
	suspensionStatuses := make(chan SuspensionStatus, status.NumRows)

	var wg sync.WaitGroup
	var userID string
	var err error

	for scanner.Scan() {
		wg.Add(1)
		userID = scanner.Text()

		// Wait for a second if number of chunked rows exceeds rowsPerChunk
		// This is to overcome RPS limitation on Cognito
		if numChunkedRows >= s.Cognito.GetRowsPerChunk() {
			time.Sleep(time.Second)
		}

		numChunkedRows++
		go s.ChanSuspend(ctx, userID, suspensionStatuses, &wg)
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

// ChanSuspend suspends with channel
func (s *Suspender) ChanSuspend(ctx context.Context, userID string, suspensionStatus chan SuspensionStatus, wg *sync.WaitGroup) {
	defer wg.Done()
	numChunkedRows--
	suspensionStatus <- SuspensionStatus{
		UserID: userID,
		Error:  s.Suspend(ctx, userID),
	}
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

		// Collect user IDs
		fmt.Fprintln(&batchBuffer.Buf, userID)
	}

	batchBuffer.NumRows = line
	return batchBuffer, nil
}

// SeqBatchSuspend suspends users in batch sequentially
func (s *Suspender) SeqBatchSuspend(ctx context.Context, buf bytes.Buffer, status BatchSuspensionStatus) (BatchSuspensionStatus, error) {
	scanner := bufio.NewScanner(&buf)
	var userID string
	var err error
	for scanner.Scan() {
		userID = scanner.Text()
		if err = s.Suspend(ctx, userID); err == nil {
			status.NumSuccessful++
			continue
		}

		// Handle error
		status.NumFailed++
		err = fmt.Errorf("failed suspending user %s: %s", userID, err.Error())
		status.Failures = append(status.Failures, err)
	}
	return status, nil
}

var numChunkedRows float32

func NewSuspender(cognito *cognito.Cognito, sqlDB *database.Database) *Suspender {
	s := new(Suspender)
	s.Cognito = cognito
	s.Database = sqlDB
	return s
}

const (
	DoneMsg = "batch suspension done, # of rows: %d, # of successful: %d, # of failed: %d\n"
)

type BatchBuffer struct {
	Buf     bytes.Buffer
	NumRows int
}

type BatchSuspensionStatus struct {
	NumRows       int
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
