package cmd

import (
	"chat-cli/internal/input_validators"
	"chat-cli/internal/lib/cli"
	"chat-cli/internal/logger"
	"context"
	descUser "github.com/gomscourse/auth/pkg/user_v1"
	"github.com/spf13/cobra"
)

// createUserCmd represents the createUser command
var createUserCmd = &cobra.Command{
	Use:   "create-user",
	Short: "Create a new user account",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		client, closer, err := getUserClient()
		if err != nil {
			logger.ErrorWithExit(err.Error())
		}

		defer closer()

		prt := &Printer{}
		username := cli.GetUserInput(
			"Username: ",
			prt,
			input_validators.NotEmpty,
		)

		password, err := cli.GetSensitiveUserInput("Password: ", prt)
		if err != nil {
			logger.ErrorWithExit(err.Error())
		}

		for {
			pc, err := cli.GetSensitiveUserInput("Confirm password: ", prt)
			if err != nil {
				logger.ErrorWithExit(err.Error())
			}

			if password != pc {
				logger.Warning("Passwords are not equal. Try again")
			} else {
				break
			}
		}

		email := cli.GetUserInput(
			"Email: ",
			prt,
			input_validators.NotEmpty,
		)

		_, err = client.Create(
			context.Background(), &descUser.CreateRequest{
				Info: &descUser.UserCreateInfo{
					Username:        username,
					Password:        password,
					PasswordConfirm: password,
					Email:           email,
					Role:            0,
				},
			},
		)

		if err != nil {
			logger.ErrorWithExit("failed to create user: %s", err)
		}

		logger.Success("user successfully created")
	},
}

func init() {
	rootCmd.AddCommand(createUserCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// createUserCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// createUserCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
