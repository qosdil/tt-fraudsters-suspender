package cmd

import (
	"context"
	"fmt"
	"log"
	"main/internal/cognito"
	"main/internal/database"
	"os"

	"github.com/brianvoe/gofakeit/v7"
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
		cognito := cognito.NewCognito(cognito.Config{
			Region: os.Getenv("AMAZON_COGNITO_CONFIG_REGION"),
			PoolID: os.Getenv("AMAZON_COGNITO_USER_POOL_ID"),
		})

		var err error
		ctx := context.Background()

		// Get Cognito client
		if cognito.Client, err = cognito.GetClient(ctx); err != nil {
			log.Fatalf("error on getting a Cognito client: %s", err.Error())
		}

		var id, email, fileContent string

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

		for i := 1; i <= numUsers; i++ {
			email = gofakeit.Email()

			if id, err = cognito.CreateUser(ctx, email); err != nil {
				log.Fatalf("error on creating a Cognito user: %s", err.Error())
			}

			if err = db.CreateUser(id, email); err != nil {
				log.Fatalf("error on creating a database user: %s", err.Error())
			}

			fileContent += id + "\n"
			fmt.Printf("#%d: %s %s\n", i, id, email)
		}

		// Remove "\n" in last line
		fileContent = fileContent[:len(fileContent)-1]

		data := []byte(fileContent)
		err = os.WriteFile(destFile, data, 0644)
		if err != nil {
			log.Fatal(err.Error())
		}
	},
}

var (
	destFile string
	numUsers int
)
