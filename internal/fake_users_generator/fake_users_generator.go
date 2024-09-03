package fake_user_generator

import (
	"context"
	"fmt"
	"main/internal/cognito"
	"main/internal/database"

	"github.com/brianvoe/gofakeit/v7"
)

type FakeUsersGenerator struct {
	Cognito  *cognito.Cognito
	Database *database.Database
}

func NewFakeUsersGenerator(cognito *cognito.Cognito, sqlDB *database.Database) *FakeUsersGenerator {
	s := new(FakeUsersGenerator)
	s.Cognito = cognito
	s.Database = sqlDB
	return s
}

func (gen *FakeUsersGenerator) CreateUser(ctx context.Context, email string) (id string, err error) {
	id, err = gen.Cognito.CreateUser(ctx, email)
	if err != nil {
		return id, err
	}
	if err := gen.Database.CreateUser(id, email); err != nil {
		// Delete user from Cognito
		if err = gen.Cognito.DeleteUser(ctx, id); err != nil {
			return id, fmt.Errorf("failed to revert user creation on Cognito: %s", err.Error())
		}

		return id, err
	}
	return id, nil
}

func (fug *FakeUsersGenerator) Generate(ctx context.Context, numUsers int) (batchText string, err error) {
	var id string
	for i := 1; i <= numUsers; i++ {
		id, err = fug.CreateUser(ctx, gofakeit.Email())
		if err != nil {
			return batchText, err
		}
		batchText += id + "\n"
	}

	// Remove "\n" in last line
	batchText = batchText[:len(batchText)-1]

	return batchText, nil
}
