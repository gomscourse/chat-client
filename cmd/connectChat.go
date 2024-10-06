package cmd

import (
	"chat-cli/internal/config"
	"chat-cli/internal/input_validators"
	"chat-cli/internal/lib/cli"
	"chat-cli/internal/logger"
	"chat-cli/internal/storage"
	"context"
	"fmt"
	descChat "github.com/gomscourse/chat-server/pkg/chat_v1"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"io"
	"slices"
	"strconv"
	"time"
)

// connectChatCmd represents the connectChat command
var connectChatCmd = &cobra.Command{
	Use:   "connect-chat",
	Short: "Connect a chat to get live messages and send yours",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		chatID, err := cmd.Flags().GetInt64("id")
		if err != nil {
			logger.ErrorWithExit("failed to get chat ID: %s", err)
		}
		st := storage.Load()
		md := metadata.MD{
			"authorization": []string{st.GetAuthHeader()},
		}

		ctx := context.Background()
		ctx = metadata.NewOutgoingContext(ctx, md)

		client, closFn, err := getChatClient()
		if err != nil {
			logger.ErrorWithExit("unable to connect to chat server")
		}

		defer closFn()

		readyCh := make(chan struct{})
		go connectChat(ctx, client, chatID, st, readyCh)
		<-readyCh
		showMessages(ctx, client, chatID, st.GetUsername())
		cli.InfinityInput(sendMessage(ctx, client, chatID), "cmd: exit")
	},
}

func init() {
	rootCmd.AddCommand(connectChatCmd)

	// Here you will define your flags and configuration settings.
	connectChatCmd.Flags().Int64("id", 0, "Chat ID to connect")

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// connectChatCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// connectChatCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func connectChat(
	ctx context.Context,
	client descChat.ChatV1Client,
	chatID int64,
	st *storage.Storage,
	readyCh chan<- struct{},
) {
	stream, err := client.ConnectChat(
		ctx, &descChat.ConnectChatRequest{
			ChatId: chatID,
		},
	)
	if err != nil {
		logger.ErrorWithExit("failed to establish stream connection: %s\n", err)
	}

	readyCh <- struct{}{}

	for {
		message, errRecv := stream.Recv()
		if errRecv == io.EOF {
			logger.ErrorWithExit("chat connection closed")
		}
		if errRecv != nil {
			logger.ErrorWithExit("failed to receive message: %s\n", errRecv)
		}

		printMessage(message, st.GetUsername())
	}
}

func printMessage(message *descChat.ChatMessage, username string) {
	author := message.GetAuthor()
	printer := logger.Simple
	if author == username {
		author = "you"
		printer = logger.Info
	}

	printer(
		"[%v] - [from: %s]: %s\n",
		message.GetCreated().AsTime().Format(time.DateTime),
		author,
		message.GetContent(),
	)
}

func sendMessage(ctx context.Context, client descChat.ChatV1Client, chatID int64) func(msg string) {
	return func(msg string) {
		switch msg {
		case "cmd: some":
			fmt.Println("cmd: some")
		default:
			fmt.Print("\033[1A")
			fmt.Print("\033[K")

			_, err := client.SendMessage(
				ctx, &descChat.SendMessageRequest{
					ChatID: chatID,
					Text:   msg,
				},
			)

			if err != nil {
				fmt.Printf("failed to send message: %s\n", err)
			}
		}
	}
}

func getChatClient() (descChat.ChatV1Client, func(), error) {
	conn, err := grpc.Dial(config.ChatServiceAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, func() {}, errors.Wrap(err, "failed to connect to server")
	}

	return descChat.NewChatV1Client(conn), func() { conn.Close() }, nil
}

func showMessages(ctx context.Context, client descChat.ChatV1Client, chatID int64, username string) {
	count := cli.GetUserInput(
		"How many last messages from this chat you want to load?",
		&Printer{},
		input_validators.IsInt,
	)
	countInt, _ := strconv.Atoi(count)
	if countInt == 0 {
		return
	}

	response, err := client.GetChatMessages(
		ctx, &descChat.GetChatMessagesRequest{
			Id:       chatID,
			Page:     1,
			PageSize: int64(countInt),
		},
	)

	if err != nil {
		logger.Warning("failed to get chat messages")
	}

	slices.SortFunc(
		response.Messages, func(a, b *descChat.ChatMessage) int {
			return int(a.GetCreated().AsTime().Unix() - b.GetCreated().AsTime().Unix())
		},
	)

	for _, msg := range response.Messages {
		printMessage(msg, username)
	}
}
