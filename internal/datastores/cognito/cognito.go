package cognito

import (
	"context"
	"fmt"
	"os"
	"strconv"
	cfg "tt_fraudsters_suspender/configs/cognito"

	AWSSDKConfig "github.com/aws/aws-sdk-go-v2/config"
	IdentityProvider "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
)

func (c *Cognito) CreateUser(ctx context.Context, email string) (id string, err error) {
	user, err := c.Client.AdminCreateUser(ctx, &IdentityProvider.AdminCreateUserInput{
		UserPoolId: &c.Config.PoolID,
		Username:   &email,
	})
	if err != nil {
		return "", err
	}

	// This returns the UUID, not email
	return *user.User.Username, nil
}

func (c *Cognito) DeleteUser(ctx context.Context, id string) (err error) {
	_, err = c.Client.AdminDeleteUser(ctx, &IdentityProvider.AdminDeleteUserInput{
		UserPoolId: &c.Config.PoolID,
		Username:   &id,
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *Cognito) DisableUser(ctx context.Context, username string) (err error) {
	if _, err = c.Client.AdminDisableUser(ctx, &IdentityProvider.AdminDisableUserInput{
		UserPoolId: &c.Config.PoolID,
		Username:   &username,
	}); err != nil {
		return err
	}
	return nil
}

func (c *Cognito) EnableUser(ctx context.Context, username string) (err error) {
	if _, err = c.Client.AdminEnableUser(ctx, &IdentityProvider.AdminEnableUserInput{
		UserPoolId: &c.Config.PoolID,
		Username:   &username,
	}); err != nil {
		return err
	}
	return nil
}

func (c *Cognito) GetClient(ctx context.Context) (*IdentityProvider.Client, error) {
	cfg, err := AWSSDKConfig.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}
	cfg.Region = c.Config.Region
	return IdentityProvider.NewFromConfig(cfg), nil
}

func (c *Cognito) GetRowsPerChunk() float32 {
	return *rowsPerChunk
}

func (c *Cognito) Truncate(ctx context.Context) (int, error) {
	lastTruncateNumAffected = 0

	// Cannot retrieve more than 60, default is 60
	users, err := c.Client.ListUsers(ctx, &IdentityProvider.ListUsersInput{
		UserPoolId: &c.Config.PoolID,
	})
	if err != nil {
		return 0, err
	}

	for _, user := range users.Users {
		_, err := c.Client.AdminDeleteUser(ctx, &IdentityProvider.AdminDeleteUserInput{
			UserPoolId: &c.Config.PoolID,
			Username:   user.Username,
		})
		if err != nil {
			return 0, err
		}
		lastTruncateNumAffected++
		truncateNumAffected++
	}

	// Recurse if there are still rows to delete
	if lastTruncateNumAffected >= 60 {
		c.Truncate(ctx)
	}

	return truncateNumAffected, nil
}

func NewCognito(config Config) (*Cognito, error) {
	c := new(Cognito)
	c.Config = config
	if err := setRowsPerChunk(); err != nil {
		return c, err
	}

	return c, nil
}

func setRowsPerChunk() (err error) {
	maxRPS64, err := strconv.ParseFloat(os.Getenv(cfg.KeyMaxRPS), 32)
	if err != nil {
		return fmt.Errorf("failed to parse env var of %s: %s", cfg.KeyMaxRPS, err.Error())
	}

	maxRPSChunkRatio64, err := strconv.ParseFloat(os.Getenv(cfg.KeyMaxRPSChunkRatio), 32)
	if err != nil {
		return fmt.Errorf("failed to parse env var of %s: %s", cfg.KeyMaxRPSChunkRatio, err.Error())
	}

	result := float32(maxRPS64) * float32(maxRPSChunkRatio64)
	rowsPerChunk = &result
	return nil
}

var (
	lastTruncateNumAffected int

	// TODO convert to *int
	rowsPerChunk *float32

	truncateNumAffected int
)

type Cognito struct {
	Config Config
	Client *IdentityProvider.Client
}

type Config struct {
	Region string
	PoolID string
}
