package cognito

import (
	"context"

	AWSSDKConfig "github.com/aws/aws-sdk-go-v2/config"
	IdentityProvider "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go/aws"
)

func (c *Cognito) CreateUser(ctx context.Context, email string) (id string, err error) {
	user, err := c.Client.AdminCreateUser(ctx, &IdentityProvider.AdminCreateUserInput{
		UserPoolId: aws.String(c.Config.PoolID),
		Username:   aws.String(email),
	})
	if err != nil {
		return "", err
	}

	// This returns the UUID, not email
	return *user.User.Username, nil
}

func (c *Cognito) DisableUser(ctx context.Context, username string) (err error) {
	if _, err = c.Client.AdminDisableUser(ctx, &IdentityProvider.AdminDisableUserInput{
		UserPoolId: aws.String(c.Config.PoolID),
		Username:   aws.String(username),
	}); err != nil {
		return err
	}
	return nil
}

func (c *Cognito) EnableUser(ctx context.Context, username string) (err error) {
	if _, err = c.Client.AdminEnableUser(ctx, &IdentityProvider.AdminEnableUserInput{
		UserPoolId: aws.String(c.Config.PoolID),
		Username:   aws.String(username),
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

func (c *Cognito) Truncate(ctx context.Context) (int, error) {
	lastTruncateNumAffected = 0

	// Cannot retrieve more than 60, default is 60
	users, err := c.Client.ListUsers(ctx, &IdentityProvider.ListUsersInput{
		UserPoolId: aws.String(c.Config.PoolID),
	})
	if err != nil {
		return 0, err
	}

	for _, user := range users.Users {
		_, err := c.Client.AdminDeleteUser(ctx, &IdentityProvider.AdminDeleteUserInput{
			UserPoolId: aws.String(c.Config.PoolID),
			Username:   aws.String(*user.Username),
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

func NewCognito(config Config) *Cognito {
	c := new(Cognito)
	c.Config = config
	return c
}

type Cognito struct {
	Config Config
	Client *IdentityProvider.Client
}

type Config struct {
	Region string
	PoolID string
}

var lastTruncateNumAffected, truncateNumAffected int
