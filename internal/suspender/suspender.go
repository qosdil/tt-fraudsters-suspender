package suspender

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"log"
	"main/internal/cognito"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
)

func (s *Suspender) BatchSuspend(ctx context.Context, buf bytes.Buffer, status BatchSuspensionStatus) BatchSuspensionStatus {
	scanner := bufio.NewScanner(&buf)
	suspensionStatuses := make(chan SuspensionStatus, status.NumRecords)

	var wg sync.WaitGroup
	var userID string
	var err error

	for scanner.Scan() {
		wg.Add(1)
		userID = scanner.Text()
		if err = s.Suspend(ctx, userID); err != nil {
			suspensionStatuses <- SuspensionStatus{UserID: userID, Error: err}
			wg.Done()
			continue
		}
		suspensionStatuses <- SuspensionStatus{UserID: userID}
		wg.Done()
	}

	go func() {
		wg.Wait()
		close(suspensionStatuses)
	}()

	for suspensionStatus := range suspensionStatuses {
		if suspensionStatus.Error != nil {
			status.NumFailed++
			err = fmt.Errorf("failed suspending user %s: %s", suspensionStatus.UserID, suspensionStatus.Error.Error())
			status.Failures = append(status.Failures, err)
			continue
		}
		status.NumSuccessful++
	}

	return status
}

func (s *Suspender) Suspend(ctx context.Context, userID string) (err error) {
	if err := s.Cognito.DisableUser(ctx, userID); err != nil {
		return err
	}
	if err := s.UpdateDatabase(userID); err != nil {
		// Re-enable user on Cognito
		s.Cognito.EnableUser(ctx, userID)

		return err
	}
	return nil
}

func (s *Suspender) CreateBufFromFile(ctx context.Context, sourceFile string) (buf bytes.Buffer, numRecords int, err error) {
	// Open the source file
	file, err := os.Open(sourceFile)
	if err != nil {
		return buf, numRecords, err
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
			return buf, numRecords, fmt.Errorf(`"%s" on line %d is not a valid UUID v4`, userID, line)
		}
		numRecords++

		// Collect user IDs
		fmt.Fprintln(&buf, userID)
	}

	return buf, numRecords, nil
}

func (s *Suspender) UpdateDatabase(userID string) error {
	log.Printf("[simulation] updating db row for user %s...\n", userID)
	if userID == "e9aa25ac-a061-70fa-0bd0-2ee61818a6b2" {
		return fmt.Errorf("failed to connect to database")
	}
	time.Sleep(50 * time.Millisecond)
	return nil
}

func NewSuspender(cognito *cognito.Cognito) *Suspender {
	s := new(Suspender)
	s.Cognito = cognito
	return s
}

const (
	DoneMsg = "batch suspension done, # of records: %d, # of successful: %d, # of failed: %d\n"
)

type BatchSuspensionStatus struct {
	NumRecords    int
	NumSuccessful int
	NumFailed     int
	Failures      []error
}

type Suspender struct {
	Cognito *cognito.Cognito
}

type SuspensionStatus struct {
	UserID string `json:"user_id"`
	Error  error  `json:"error"`
}
