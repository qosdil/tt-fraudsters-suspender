package cognito

import (
	"context"

	AWSSDKConfig "github.com/aws/aws-sdk-go-v2/config"
	IdentityProvider "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go/aws"
)

type Config struct {
	Region string
	PoolID string
}

type Cognito struct {
	Config Config
	Client *IdentityProvider.Client
}

func NewCognito(config Config) *Cognito {
	c := new(Cognito)
	c.Config = config
	return c
}

func (c *Cognito) GetClient(ctx context.Context) (*IdentityProvider.Client, error) {
	cfg, err := AWSSDKConfig.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}
	cfg.Region = c.Config.Region
	return IdentityProvider.NewFromConfig(cfg), nil
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
