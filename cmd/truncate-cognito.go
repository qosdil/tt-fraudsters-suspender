/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log"

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
		log.Printf("start truncating Cognito user pool %s...", cognitoConn.Config.PoolID)
		numAffected, err := cognitoConn.Truncate(ctx)
		if err != nil {
			log.Fatalf("error on truncating Cognito user pool: %s", err.Error())
		}

		if numAffected > 0 {
			log.Printf("successfully deleted %d Cognito user(s)\n", numAffected)
		} else {
			log.Println("user pool is empty")
		}
	},
}

func init() {
	rootCmd.AddCommand(truncateCognitoCmd)
}
