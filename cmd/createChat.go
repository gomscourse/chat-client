package cmd

import (
	"chat-cli/internal/input_validators"
	"chat-cli/internal/lib/cli"
	"chat-cli/internal/logger"
	"chat-cli/internal/storage"
	"context"
	descChat "github.com/gomscourse/chat-server/pkg/chat_v1"
	"github.com/gomscourse/common/pkg/sys/messages"
	"github.com/gomscourse/common/pkg/tools"
	"github.com/pkg/errors"
	"strings"

	"github.com/spf13/cobra"
)

// createChatCmd represents the createChat command
var createChatCmd = &cobra.Command{
	Use:   "create-chat",
	Short: "Create a new chat and add users to it",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		st := storage.Load()
		ctx := context.Background()
		ctx = getRequestContext(ctx, st)

		client, closFn, err := getChatClient()
		if err != nil {
			logger.ErrorWithExit("unable to connect to chat server")
		}

		defer closFn()

		prt := &Printer{}
		title := cli.GetUserInput(
			"Chat title: ",
			prt,
			input_validators.NotEmpty,
		)

		usernames := cli.GetUserInput(
			"Usernames, separated by comma: ",
			prt,
			input_validators.NotEmpty,
		)

		usernamesSlice := strings.Split(usernames, ",")

		res, err := client.Create(
			ctx, &descChat.CreateRequest{
				Title: title,
				Usernames: tools.MapSlice(
					usernamesSlice, func(u string) string {
						return strings.TrimSpace(u)
					},
				),
			},
		)

		if err != nil {
			var se GRPCStatusInterface
			if errors.As(err, &se) && se.GRPCStatus().Message() == messages.AccessTokenInvalid {
				refreshAccessToken(ctx, st)
				ctx = getRequestContext(ctx, st)
				res, err = client.Create(
					ctx, &descChat.CreateRequest{
						Title: title,
						Usernames: tools.MapSlice(
							usernamesSlice, func(u string) string {
								return strings.TrimSpace(u)
							},
						),
					},
				)
				if err != nil {
					logger.ErrorWithExit("failed to create chat: %s", err.Error())
				}
			} else {
				logger.ErrorWithExit("failed to create chat: %s", err.Error())
			}
		}

		logger.Success("Created chat with id %d", res.Id)
	},
}

func init() {
	rootCmd.AddCommand(createChatCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// createChatCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// createChatCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
