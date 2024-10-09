package cmd

import (
	"log"
	"os"
	"time"
	generator "tt-fraudsters-suspender/internal/pkg/fake_users_generator"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(generateFakeUsersCmd)
	generateFakeUsersCmd.Flags().IntVarP(&numUsers, "num-users", "n", 0, "Number of users to be generated (required)")
	generateFakeUsersCmd.MarkFlagRequired("num-users")
	generateFakeUsersCmd.Flags().StringVarP(&destFile, "dest-file", "d", "", "Destination file (required)")
	generateFakeUsersCmd.MarkFlagRequired("dest-file")
}

// generateFakeUsersCmd represents the generateFakeUsers command
var generateFakeUsersCmd = &cobra.Command{
	Use:   "generate-fake-users",
	Short: "Generate rows of fake users in a text file, also insert them to Cognito and into the database",
	Long: `Generate rows of fake users in a text file, also insert them to Cognito and into the database.

Command example:

fraudster_suspender generate-fake-users --num-users=1000 --dest-file=$HOME/Downloads/fraudsters.txt`,
	Run: func(cmd *cobra.Command, args []string) {
		start := time.Now()
		setDatabaseConnection()
		defer dbConn.SqlDB.Close()

		log.Printf("start generating %d fake users...", numUsers)
		gen := generator.NewFakeUsersGenerator(cognitoConn, dbConn)
		batchText, err := gen.Generate(ctx, numUsers)
		if err != nil {
			log.Fatalf("error on generating fake users: %s", err.Error())
		}

		data := []byte(batchText)
		err = os.WriteFile(destFile, data, 0644)
		if err != nil {
			log.Fatal(err.Error())
		}
		elapsed := time.Since(start).Seconds()

		log.Printf("successfully generated %d fake users to Cognito, database and batch text file", numUsers)
		log.Printf("done in %.2fs\n", elapsed)
	},
}

var (
	destFile string
	numUsers int
)
