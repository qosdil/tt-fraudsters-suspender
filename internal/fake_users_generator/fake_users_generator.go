package fake_user_generator

import (
	"context"
	"fmt"
	"main/internal/cognito"
	"main/internal/database"
	"os"
	"strconv"
	"sync"
	"time"

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

	rowsPerChunk, err := gen.getRowsPerChunk()
	if err != nil {
		return batchText, err
	}

	for i := 1; i <= numUsers; i++ {
		wg.Add(1)

		// Wait for a second if number of chunked rows exceeds rowsPerChunk
		// This is to overcome RPS limitation on Cognito
		if numChunkRows >= rowsPerChunk {
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

func (gen *FakeUsersGenerator) getRowsPerChunk() (rowsPerChunk float32, err error) {
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

func NewFakeUsersGenerator(cognito *cognito.Cognito, sqlDB *database.Database) *FakeUsersGenerator {
	s := new(FakeUsersGenerator)
	s.Cognito = cognito
	s.Database = sqlDB
	return s
}

var numChunkRows float32

const (
	envCognitoMaxRPS           = "AMAZON_COGNITO_MAX_RPS"
	envCognitoMaxRPSChunkRatio = "AMAZON_COGNITO_MAX_RPS_CHUNK_RATIO"
)

type CreationStatus struct {
	ID    string
	Error error
}

type FakeUsersGenerator struct {
	Cognito  *cognito.Cognito
	Database *database.Database
}
