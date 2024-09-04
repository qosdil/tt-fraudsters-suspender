package cmd

import (
	"context"
	"log"
	"main/internal/cognito"
	"main/internal/database"
	generator "main/internal/fake_users_generator"
	"os"
	"time"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(generateFakeUsersCmd)
	generateFakeUsersCmd.Flags().IntVarP(&numUsers, "num-users", "n", 0, "Number of users to be generated")
	generateFakeUsersCmd.Flags().StringVarP(&destFile, "dest-file", "d", "", "Destination file")
}

// generateFakeUsersCmd represents the generateFakeUsers command
var generateFakeUsersCmd = &cobra.Command{
	Use:   "generate-fake-users",
	Short: "Generate a text file, insert data to Cognito and database",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		start := time.Now()
		var err error
		cognito, err := cognito.NewCognito(cognito.Config{
			Region: os.Getenv("AMAZON_COGNITO_CONFIG_REGION"),
			PoolID: os.Getenv("AMAZON_COGNITO_USER_POOL_ID"),
		})
		if err != nil {
			log.Fatalf("error on instantiating Cognito: %s", err.Error())
		}

		ctx := context.Background()

		// Get Cognito client
		if cognito.Client, err = cognito.GetClient(ctx); err != nil {
			log.Fatalf("error on getting a Cognito client: %s", err.Error())
		}

		// Create a database connection
		db := database.NewDatabase(database.Config{
			Host:     os.Getenv("DB_HOST"),
			Port:     os.Getenv("DB_PORT"),
			User:     os.Getenv("DB_USER"),
			Password: os.Getenv("DB_PASSWORD"),
			Name:     os.Getenv("DB_NAME"),
			SSLMode:  os.Getenv("DB_SSL_MODE"),
		})
		if db.SqlDB, err = db.Open(); err != nil {
			log.Fatalf("error on opening a connection to database: %s", err.Error())
		}
		defer db.SqlDB.Close()

		log.Printf("start generating %d fake users...", numUsers)
		gen := generator.NewFakeUsersGenerator(cognito, db)
		batchText, err := gen.Generate(ctx, numUsers)
		if err != nil {
			log.Fatalf("error on generating fake users: %s", err.Error())
		}

		data := []byte(batchText)
		err = os.WriteFile(destFile, data, 0644)
		if err != nil {
			log.Fatal(err.Error())
		}
		elapsed := time.Since(start)

		log.Printf("successfully generated %d fake users to Cognito, database and batch text file", numUsers)
		log.Printf("done in %s\n", elapsed)
	},
}

var (
	destFile string
	numUsers int
)
