package cmd

import (
	"log"
	"time"
	"tt-fraudsters-suspender/internal/pkg/suspender"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(seqSuspendCmd)
	seqSuspendCmd.Flags().StringVarP(&sourceFile, "source-file", "s", "", "Source file to read from (required)")
	seqSuspendCmd.MarkFlagRequired("source-file")
}

var seqSuspendCmd = &cobra.Command{
	Use:   "seq-suspend",
	Short: "Sequentially suspend users based on the IDs that you provide in <source-file> file",
	Long: `Sequentially suspend users based on the IDs that you provide in <source-file> file.

The following is the example of source file content with 3 users to suspend:

6f8c96fc-fde9-4af6-8f09-18144c6cf278
39d1a303-ca47-432e-96fe-e1e9ccba0c6a
83525ffb-15d5-4d04-a517-ce830b3f77a9

Command example:

fraudster_suspender seq-suspend --source-file=/Users/john/Downloads/fraudsters.txt`,
	Run: func(cmd *cobra.Command, args []string) {
		start := time.Now()
		setDatabaseConnection()
		defer dbConn.SqlDB.Close()

		s := suspender.NewSuspender(cognitoConn, dbConn)
		batchBuffer, err := s.CreateBufFromFile(ctx, sourceFile)
		if err != nil {
			log.Fatal(err.Error())
		}

		log.Printf("start suspending %d users...", batchBuffer.NumRows)
		batchStatus, err := s.SeqBatchSuspend(ctx, batchBuffer.Buf, suspender.BatchSuspensionStatus{
			NumRows: batchBuffer.NumRows,
		})
		if err != nil {
			log.Fatalf("failed to suspend in batch: %s", err.Error())
		}
		elapsed := time.Since(start).Seconds()

		// Output failures if there's any
		for _, failure := range batchStatus.Failures {
			log.Println(failure.Error())
		}

		log.Printf(suspender.DoneMsg, batchStatus.NumRows, batchStatus.NumSuccessful, batchStatus.NumFailed)
		log.Printf("done in %.2fs\n", elapsed)
	},
}
