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
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
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
	var err error
	dbConn = database.NewDatabase(database.Config{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		Name:     os.Getenv("DB_NAME"),
		SSLMode:  os.Getenv("DB_SSL_MODE"),
	})
	if dbConn.SqlDB, err = dbConn.Open(); err != nil {
		log.Fatalf("error on opening a connection to database: %s", err.Error())
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
