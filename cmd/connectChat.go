package cmd

import (
	"chat-cli/internal/config"
	"chat-cli/internal/lib/cli"
	"chat-cli/internal/storage"
	"context"
	"fmt"
	descChat "github.com/gomscourse/chat-server/pkg/chat_v1"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"io"
	"log"
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
			log.Fatalf("failed to get chat ID: %s", err)
		}

		readyCh := make(chan struct{})
		go connectChat(chatID, readyCh)
		<-readyCh
		cli.InfinityInput(sendMessage, "cmd: exit")
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

func connectChat(chatID int64, readyCh chan<- struct{}) {
	st := storage.Load()
	md := metadata.MD{
		"authorization": []string{st.GetAuthHeader()},
	}
	conn, err := grpc.Dial(config.ChatServiceAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("failed to connect to server: %s\n", err)
		return
	}
	defer conn.Close()

	ctx := context.Background()
	client := descChat.NewChatV1Client(conn)

	ctx = metadata.NewOutgoingContext(ctx, md)
	stream, err := client.ConnectChat(
		ctx, &descChat.ConnectChatRequest{
			ChatId: chatID,
		},
	)
	if err != nil {
		log.Printf("failed to establish stream connection: %s\n", err)
		return
	}

	readyCh <- struct{}{}

	for {
		message, errRecv := stream.Recv()
		if errRecv == io.EOF {
			log.Printf("chat connection closed")
			return
		}
		if errRecv != nil {
			log.Printf("failed to receive message: %s\n", errRecv)
			return
		}

		author := message.GetAuthor()
		if author == st.GetUsername() {
			author = "you"
		}

		fmt.Printf(
			"[%v] - [from: %s]: %s\n",
			message.GetCreated().AsTime().Format(time.RFC3339),
			author,
			message.GetContent(),
		)
	}
}

func sendMessage(msg string) {
	switch msg {
	case "cmd: some":
		fmt.Println("cmd: some")
	default:
		fmt.Print("\033[1A")
		fmt.Print("\033[K")

		//TODO: отправка сообщения
		fmt.Println(fmt.Sprintf("from you: %s", msg))
	}
}
