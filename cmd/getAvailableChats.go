package cmd

import (
	"chat-cli/internal/input_validators"
	"chat-cli/internal/lib/cli"
	"chat-cli/internal/logger"
	"chat-cli/internal/storage"
	"context"
	"fmt"
	descChat "github.com/gomscourse/chat-server/pkg/chat_v1"
	"github.com/gomscourse/common/pkg/sys/messages"
	"github.com/pkg/errors"
	"strconv"

	"github.com/spf13/cobra"
)

// getAvailableChatsCmd represents the getAvailableChats command
var getAvailableChatsCmd = &cobra.Command{
	Use:   "get-available-сhats",
	Short: "Get chats available for logged in user",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		client, closFn, err := getChatClient()
		if err != nil {
			logger.ErrorWithExit("unable to connect to chat server")
		}

		defer closFn()

		st := storage.Load()
		ctx := getRequestContext(st)

		count := cli.GetUserInput(
			"How many chats you want to load (empty to load all)?",
			&Printer{},
			input_validators.IsIntOrEmpty,
		)
		countInt := 0
		if count != "" {
			ci, _ := strconv.Atoi(count)
			countInt = ci
		}

		response, err := getChats(ctx, client, countInt)
		if err != nil {
			var se GRPCStatusInterface
			if errors.As(err, &se) && se.GRPCStatus().Message() == messages.AccessTokenInvalid {
				refreshAccessToken(ctx, st)
				response, err = getChats(getRequestContext(st), client, countInt)
				if err != nil {
					logger.ErrorWithExit("failed to get available chats: %s", err)
				}
			} else {
				logger.ErrorWithExit("failed to get available chats: %s", err)
			}
		}

		if len(response.Chats) == 0 {
			logger.Warning("You have not been added to any chat yet")
			return
		}

		printAvailableChats(response.Chats)
	},
}

func init() {
	rootCmd.AddCommand(getAvailableChatsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// getAvailableChatsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// getAvailableChatsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func getChats(ctx context.Context, client descChat.ChatV1Client, count int) (
	*descChat.GetAvailableChatsResponse,
	error,
) {
	return client.GetAvailableChats(
		ctx, &descChat.GetAvailableChatsRequest{
			Page:     1,
			PageSize: int64(count),
		},
	)
}

func printAvailableChats(chats []*descChat.Chat) {
	fmt.Printf("%-5s %-20s %-20s\n", "ID", "Title", "Created")
	fmt.Println("------------------------------------------------------------")

	for _, item := range chats {
		fmt.Printf("%-5d %-20s %-20s\n", item.ID, item.Title, item.Created.AsTime().Format("2006-01-02 15:04:05"))
	}

	fmt.Println("------------------------------------------------------------")
}
