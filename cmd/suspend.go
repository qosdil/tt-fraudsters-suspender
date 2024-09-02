package cmd

import (
	"context"
	"log"
	"main/internal/cognito"
	susp "main/internal/suspender"
	"os"

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
		cognito.Client, err = cognito.GetClient(ctx)
		if err != nil {
			log.Fatalf("error on getting a Cognito client: %s", err.Error())
		}

		suspender := susp.NewSuspender(cognito)
		buf, numRecords, err := suspender.CreateBufFromFile(ctx, sourceFile)
		if err != nil {
			log.Fatal(err.Error())
		}

		batchStatus := suspender.BatchSuspend(ctx, buf, susp.BatchSuspensionStatus{
			NumRecords: numRecords,
		})

		// Output failures if there's any
		for _, failure := range batchStatus.Failures {
			log.Println(failure.Error())
		}

		log.Printf(susp.DoneMsg, batchStatus.NumRecords, batchStatus.NumSuccessful, batchStatus.NumFailed)
	},
}
