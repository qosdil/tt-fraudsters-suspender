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

func (fug *FakeUsersGenerator) Generate(ctx context.Context, numUsers int) (batchText string, err error) {
	var id, email string
	for i := 1; i <= numUsers; i++ {
		email = gofakeit.Email()
		if id, err = fug.Cognito.CreateUser(ctx, email); err != nil {
			return "", fmt.Errorf("error on creating a Cognito user: %s", err.Error())
		}
		if err = fug.Database.CreateUser(id, email); err != nil {
			return "", fmt.Errorf("error on creating a database user: %s", err.Error())
		}
		batchText += id + "\n"
	}

	// Remove "\n" in last line
	batchText = batchText[:len(batchText)-1]

	return batchText, nil
}
