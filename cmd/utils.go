package cmd

import (
	"chat-cli/internal/config"
	"chat-cli/internal/storage"
	"context"
	descChat "github.com/gomscourse/chat-server/pkg/chat_v1"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

func getChatClient() (descChat.ChatV1Client, func(), error) {
	conn, err := grpc.Dial(config.ChatServiceAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, func() {}, errors.Wrap(err, "failed to connect to server")
	}

	return descChat.NewChatV1Client(conn), func() { conn.Close() }, nil
}

func getRequestContext(st *storage.Storage) context.Context {
	md := metadata.MD{
		"authorization": []string{st.GetAuthHeader()},
	}

	ctx := context.Background()
	return metadata.NewOutgoingContext(ctx, md)
}
