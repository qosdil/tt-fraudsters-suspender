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
		elapsed := time.Since(start)

		log.Printf("successfully generated %d fake users to Cognito, database and batch text file", numUsers)
		log.Printf("done in %s\n", elapsed)
	},
}

var (
	destFile string
	numUsers int
)
