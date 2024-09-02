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

func (s *Suspender) SuspendFromFile(ctx context.Context, sourceFile string) (err error) {
	// Open the source file
	readFile, err := os.Open(sourceFile)
	if err != nil {
		return err
	}
	defer readFile.Close()

	// Scan the content
	fileScanner := bufio.NewScanner(readFile)
	fileScanner.Split(bufio.ScanLines)

	var userID string

	// For the rescan
	buf := bytes.Buffer{}

	// Count number of lines, validate UUID v4
	line := 1
	for fileScanner.Scan() {
		userID = fileScanner.Text()
		fmt.Fprintln(&buf, userID)
		if _, err := uuid.Parse(userID); err != nil {
			return fmt.Errorf(`"%s" on line %d is not a valid UUID v4`, userID, line)
		}
		line++
	}

	// Rescan
	fileScanner = bufio.NewScanner(&buf)

	batchStatus := BatchSuspensionStatus{
		NumRecords: line,
	}
	suspensionStatuses := make(chan SuspensionStatus, batchStatus.NumRecords)
	var wg sync.WaitGroup
	for fileScanner.Scan() {
		wg.Add(1)
		userID = fileScanner.Text()
		log.Printf("suspending user ID %s...\n", userID)
		if err := s.Suspend(ctx, userID); err != nil {
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
			batchStatus.NumFailed++
			log.Printf("failed suspending user ID %s: %s\n", suspensionStatus.UserID, suspensionStatus.Error.Error())
			continue
		}

		batchStatus.NumSuccessful++
		log.Printf("successful suspending user ID %s\n", suspensionStatus.UserID)
	}

	log.Printf(doneMsg, batchStatus.NumRecords, batchStatus.NumSuccessful, batchStatus.NumFailed)
	return nil
}

func (s *Suspender) UpdateDatabase(userID string) error {
	log.Printf("[simulation] updating db row for user ID %s...\n", userID)
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
	doneMsg = "batch suspension done, # of records: %d, # of successful: %d, # of failed: %d\n"
)

type BatchSuspensionStatus struct {
	NumRecords    int
	NumSuccessful int
	NumFailed     int
}

type Suspender struct {
	Cognito *cognito.Cognito
}

type SuspensionStatus struct {
	UserID string `json:"user_id"`
	Error  error  `json:"error"`
}
