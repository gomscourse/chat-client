package cmd

import (
	"chat-cli/internal/config"
	"chat-cli/internal/storage"
	"context"
	descAuth "github.com/gomscourse/auth/pkg/auth_v1"
	descUser "github.com/gomscourse/auth/pkg/user_v1"
	descChat "github.com/gomscourse/chat-server/pkg/chat_v1"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"log"
)

func getChatClient() (descChat.ChatV1Client, func(), error) {
	conn, err := grpc.Dial(config.ChatServiceAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, func() {}, errors.Wrap(err, "failed to connect to server")
	}

	return descChat.NewChatV1Client(conn), func() { conn.Close() }, nil
}

func getAuthClient() (descAuth.AuthV1Client, func(), error) {
	creds, err := credentials.NewClientTLSFromFile("service.pem", "")
	if err != nil {
		log.Fatalf("could not process the credentials: %v", err)
	}

	conn, err := grpc.Dial(
		config.AuthServiceAddress,
		grpc.WithTransportCredentials(creds),
	)

	if err != nil {
		return nil, func() {}, errors.Wrap(err, "failed to connect to auth server")
	}

	return descAuth.NewAuthV1Client(conn), func() { conn.Close() }, nil
}

func getUserClient() (descUser.UserV1Client, func(), error) {
	conn, err := grpc.Dial(config.AuthServiceAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, func() {}, errors.Wrap(err, "failed to connect to auth server")
	}

	return descUser.NewUserV1Client(conn), func() { conn.Close() }, nil
}

func getRequestContext(st *storage.Storage) context.Context {
	md := metadata.MD{
		"authorization": []string{st.GetAuthHeader()},
	}

	ctx := context.Background()
	return metadata.NewOutgoingContext(ctx, md)
}
