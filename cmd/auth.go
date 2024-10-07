package cmd

import (
	"chat-cli/internal/input_validators"
	"chat-cli/internal/lib/cli"
	"chat-cli/internal/logger"
	"chat-cli/internal/storage"
	"context"
	"fmt"
	descAuth "github.com/gomscourse/auth/pkg/auth_v1"
	"github.com/spf13/cobra"
	"log"
)

type Printer struct{}

func (p *Printer) Info(msg string) {
	fmt.Println(msg)
}
func (p *Printer) Warning(msg string) {
	fmt.Println(msg)
}

// authCmd represents the auth command
var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Login to the app",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		printer := &Printer{}
		username := cli.GetUserInput("Username: ", printer, input_validators.NotEmpty)
		password, err := cli.GetSensitiveUserInput("Password", printer)
		if err != nil {
			log.Fatalf(fmt.Sprintf("Failed to read password: %s", err.Error()))
		}

		authClient, closer, err := getAuthClient()
		if err != nil {
			logger.ErrorWithExit(err.Error())
		}

		defer closer()

		ctx := context.Background()
		loginPayload := &descAuth.LoginRequest{
			Username: username,
			Password: password,
		}

		loginRes, err := authClient.Login(ctx, loginPayload)
		if err != nil {
			log.Fatalf("failed to get refresh token: %v", err)
		}
		refreshToken := loginRes.GetRefreshToken()

		st := storage.Load()
		st.SetRefreshToken(refreshToken)

		atPayload := &descAuth.GetAccessTokenRequest{RefreshToken: refreshToken}
		atRes, err := authClient.GetAccessToken(ctx, atPayload)
		if err != nil {
			log.Fatalf("failed to get access token: %v", err)
		}
		st.SetUsername(username)
		st.SetAccessToken(atRes.GetAccessToken())
		st.Flush()

		logger.Success("Successfully logged in as %s", username)
	},
}

func init() {
	rootCmd.AddCommand(authCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// authCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// authCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
