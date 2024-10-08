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
	Short: "Delete all users data on the targeted Cognito user pool",
	Long: `Delete all users data on the targeted Cognito user pool.

Command example:

fraudsters-suspender truncate-cognito`,
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
