package cmd

import (
	"chat-cli/internal/config"
	"chat-cli/internal/lib/cli"
	"context"
	"fmt"
	descAuth "github.com/gomscourse/auth/pkg/auth_v1"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
)

type Printer struct{}

func (p *Printer) Info(msg string, a ...any) {
	fmt.Println(msg)
}
func (p *Printer) Warning(msg string, a ...any) {
	fmt.Println(msg)
}

// authCmd represents the auth command
var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Log in chat app",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		printer := &Printer{}
		username := cli.GetUserInput("Username: ", printer)
		password, err := cli.GetSensitiveUserInput("Password", printer)
		if err != nil {
			log.Fatalf(fmt.Sprintf("Failed to read password: %s", err.Error()))
		}

		conn, err := grpc.Dial(config.AuthServiceAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("failed to connect to auth server: %v", err)
		}
		defer conn.Close()

		authClient := descAuth.NewAuthV1Client(conn)
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
		atPayload := &descAuth.GetAccessTokenRequest{RefreshToken: refreshToken}
		atRes, err := authClient.GetAccessToken(ctx, atPayload)
		if err != nil {
			log.Fatalf("failed to get access token: %v", err)
		}
		// TODO: сохранить токен в переменные среды или файл
		fmt.Println(atRes.GetAccessToken())
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
