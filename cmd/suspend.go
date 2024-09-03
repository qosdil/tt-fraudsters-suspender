package cmd

import (
	"context"
	"log"
	"main/internal/cognito"
	"main/internal/database"
	susp "main/internal/suspender"
	"os"
	"time"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(suspendCmd)
	suspendCmd.Flags().StringVarP(&sourceFile, "source-file", "s", "", "Source file to read from")
}

var sourceFile string

var suspendCmd = &cobra.Command{
	Use:   "suspend",
	Short: "Suspend users based on the IDs that you provide in <source-file> file",
	Long: `Suspend users based on the IDs that you provide in <source-file> file.

The following is the example of source file content with 3 users to suspend:

6f8c96fc-fde9-4af6-8f09-18144c6cf278
39d1a303-ca47-432e-96fe-e1e9ccba0c6a
83525ffb-15d5-4d04-a517-ce830b3f77a9

Command example:

fraudster_suspender suspend --source-file=/Users/john/Downloads/fraudsters.txt`,
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

		suspender := susp.NewSuspender(cognito, db)
		batchBuffer, err := suspender.CreateBufFromFile(ctx, sourceFile)
		if err != nil {
			log.Fatal(err.Error())
		}

		log.Printf("start suspending %d users...", batchBuffer.NumRecords)
		start := time.Now()
		batchStatus, err := suspender.BatchSuspend(ctx, batchBuffer.Buf, susp.BatchSuspensionStatus{
			NumRecords: batchBuffer.NumRecords,
		})
		if err != nil {
			log.Fatalf("failed to suspend in batch: %s", err.Error())
		}
		elapsed := time.Since(start)

		// Output failures if there's any
		for _, failure := range batchStatus.Failures {
			log.Println(failure.Error())
		}

		log.Printf(susp.DoneMsg, batchStatus.NumRecords, batchStatus.NumSuccessful, batchStatus.NumFailed)
		log.Printf("done in %s\n", elapsed)
	},
}
