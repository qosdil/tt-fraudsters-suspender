/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"log"
	"main/internal/cognito"
	"os"

	"github.com/spf13/cobra"
)

// truncateCognitoCmd represents the truncateCognito command
var truncateCognitoCmd = &cobra.Command{
	Use:   "truncate-cognito",
	Short: "A brief description of your command",
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

		log.Printf("start truncating Cognito user pool %s...", cognito.Config.PoolID)
		numAffected, err := cognito.Truncate(ctx)
		if err != nil {
			log.Fatalf("error on truncating Cognito user pool: %s", err.Error())
		}

		if numAffected > 0 {
			log.Printf("successfully deleted %d Cognito user(s)\n", numAffected)
		} else {
			fmt.Println("user pool is empty")
		}
	},
}

func init() {
	rootCmd.AddCommand(truncateCognitoCmd)
}
