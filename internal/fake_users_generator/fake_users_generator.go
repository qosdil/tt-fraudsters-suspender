package fake_user_generator

import (
	"context"
	"fmt"
	"main/internal/cognito"
	"main/internal/database"

	"github.com/brianvoe/gofakeit/v7"
)

func (gen *FakeUsersGenerator) CreateUser(ctx context.Context, email string) (id string, err error) {
	id, err = gen.Cognito.CreateUser(ctx, email)
	if err != nil {
		return id, err
	}
	if err = gen.Database.CreateUser(id, email); err != nil {
		// Delete user from Cognito
		if duErr := gen.Cognito.DeleteUser(ctx, id); err != nil {
			return id, fmt.Errorf("failed to revert user creation on Cognito: %s", duErr.Error())
		}

		return id, err
	}
	return id, nil
}

func (gen *FakeUsersGenerator) Generate(ctx context.Context, numUsers int) (batchText string, err error) {
	var id string
	for i := 1; i <= numUsers; i++ {
		id, err = gen.CreateUser(ctx, gofakeit.Email())
		if err != nil {
			return batchText, err
		}
		batchText += id + "\n"
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

type FakeUsersGenerator struct {
	Cognito  *cognito.Cognito
	Database *database.Database
}
