package fake_user_generator

import (
	"context"
	"fmt"
	"sync"
	"time"
	"tt-fraudsters-suspender/internal/datastores/cognito"
	database "tt-fraudsters-suspender/internal/datastores/postgres"

	"github.com/brianvoe/gofakeit/v7"
)

// ChanCreateUser creates user with channel
func (gen *FakeUsersGenerator) ChanCreateUser(ctx context.Context, email string, creationStatuses chan CreationStatus, wg *sync.WaitGroup) (err error) {
	defer wg.Done()
	id, err := gen.CreateUser(ctx, email)
	if err != nil {
		creationStatuses <- CreationStatus{ID: id, Error: err}
		numChunkRows--
		return err
	}

	creationStatuses <- CreationStatus{ID: id}
	numChunkRows--
	return nil
}

func (gen *FakeUsersGenerator) CreateUser(ctx context.Context, email string) (id string, err error) {
	id, err = gen.Cognito.CreateUser(ctx, email)
	if err != nil {
		return id, err
	}
	if err = gen.Database.CreateUser(id, email); err != nil {
		// Delete user from Cognito
		if duErr := gen.Cognito.DeleteUser(ctx, id); duErr != nil {
			return id, fmt.Errorf("failed to revert user creation on Cognito: %s", duErr.Error())
		}

		return id, err
	}
	return id, nil
}

func (gen *FakeUsersGenerator) Generate(ctx context.Context, numUsers int) (batchText string, err error) {
	creationStatuses := make(chan CreationStatus, numUsers)
	var wg sync.WaitGroup

	for i := 1; i <= numUsers; i++ {
		wg.Add(1)

		// Wait for a second if number of chunked rows exceeds rowsPerChunk
		// This is to overcome RPS limitation on Cognito
		if numChunkRows >= gen.Cognito.GetRowsPerChunk() {
			time.Sleep(time.Second)
		}

		numChunkRows++
		go gen.ChanCreateUser(ctx, gofakeit.Email(), creationStatuses, &wg)
	}

	go func() {
		wg.Wait()
		close(creationStatuses)
	}()

	for creationStatus := range creationStatuses {
		if creationStatus.Error == nil {
			batchText += creationStatus.ID + "\n"
			continue
		}

		return batchText, err
	}

	// Remove "\n" in last line
	batchText = batchText[:len(batchText)-1]

	return batchText, nil
}

func NewFakeUsersGenerator(cognito *cognito.Cognito, sqlDB *database.Database) *FakeUsersGenerator {
	s := new(FakeUsersGenerator)
	s.Cognito = cognito
	s.Database = sqlDB
	return s
}

var numChunkRows float32

type CreationStatus struct {
	ID    string
	Error error
}

type FakeUsersGenerator struct {
	Cognito  *cognito.Cognito
	Database *database.Database
}
