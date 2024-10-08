package cmd

import (
	"context"
	"log"
	"os"
	"tt-fraudsters-suspender/internal/datastores/cognito"
	database "tt-fraudsters-suspender/internal/datastores/postgres"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "main",
	Short: "Demonstrates the fast batch update processes to Amazon Cognito and PostgreSQL database simultaneously",
	Long:  `A collection of CLI programs that demonstrate the fast batch update processes to Amazon Cognito and PostgreSQL database simultaneously by leveraging Go's concurrency.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	// Set "cognitoConn" as a Cognito connection in "cmd" package
	ctx = context.Background()
	setCognitoConnection()

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

var (
	cognitoConn *cognito.Cognito
	ctx         context.Context
	dbConn      *database.Database
)

// setCognitoConnection returns a Cognito connection
func setCognitoConnection() {
	var err error
	cognitoConn, err = cognito.NewCognito(cognito.Config{
		Region: os.Getenv("AMAZON_COGNITO_CONFIG_REGION"),
		PoolID: os.Getenv("AMAZON_COGNITO_USER_POOL_ID"),
	})
	if err != nil {
		log.Fatalf("error on instantiating Cognito: %s", err.Error())
	}

	// Get Cognito client
	if cognitoConn.Client, err = cognitoConn.GetClient(ctx); err != nil {
		log.Fatalf("error on getting a Cognito client: %s", err.Error())
	}
}

func setDatabaseConnection() {
	dbConn = database.NewDatabase(database.Config{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		Name:     os.Getenv("DB_NAME"),
		SSLMode:  os.Getenv("DB_SSL_MODE"),
	})

	var err error
	errMsg := "error on opening a connection to database: %s"

	dbConn.SqlDB, err = dbConn.Open()
	if err != nil {
		log.Fatalf(errMsg, err.Error())
	}
	err = dbConn.SqlDB.Ping()
	if err != nil {
		log.Fatalf(errMsg, err.Error())
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.main.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
