package cmd

import (
	"chat-cli/internal/config"
	"chat-cli/internal/logger"
	"chat-cli/internal/storage"
	"context"
	"crypto/tls"
	"crypto/x509"
	_ "embed"
	descAuth "github.com/gomscourse/auth/pkg/auth_v1"
	descUser "github.com/gomscourse/auth/pkg/user_v1"
	descChat "github.com/gomscourse/chat-server/pkg/chat_v1"
	"github.com/gomscourse/common/pkg/sys/messages"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"log"
	"os"
)

func getChatClient() (descChat.ChatV1Client, func(), error) {
	conn, err := grpc.Dial(config.ChatServiceAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, func() {}, errors.Wrap(err, "failed to connect to server")
	}

	return descChat.NewChatV1Client(conn), func() { conn.Close() }, nil
}

//go:embed cert/cert.pem
var certData []byte

func getAuthClient() (descAuth.AuthV1Client, func(), error) {
	certPool := x509.NewCertPool()

	if !certPool.AppendCertsFromPEM(certData) {
		log.Fatal("could not process the credentials")
	}

	tlsConfig := &tls.Config{
		RootCAs: certPool,
	}

	conn, err := grpc.Dial(
		config.AuthServiceAddress,
		grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)),
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

func getRequestContext(ctx context.Context, st *storage.Storage) context.Context {
	md := metadata.MD{
		"authorization": []string{st.GetAuthHeader()},
	}

	return metadata.NewOutgoingContext(ctx, md)
}

func refreshAccessToken(ctx context.Context, st *storage.Storage) {
	authClient, closer, err := getAuthClient()
	if err != nil {
		logger.ErrorWithExit(err.Error())
	}

	defer closer()

	refreshToken := st.GetRefreshToken()
	aRes, err := requestAccessToken(ctx, authClient, refreshToken)

	if err != nil {
		var se GRPCStatusInterface
		if errors.As(err, &se) && se.GRPCStatus().Message() == messages.RefreshTokenInvalid {
			logger.Warning("You need to log in again")
			os.Exit(0)
		} else {
			logger.ErrorWithExit("failed to authenticate request: %s", err)
		}
	}

	st.SetRefreshToken(refreshToken)
	st.SetAccessToken(aRes.GetAccessToken())
	st.Flush()
}

func requestAccessToken(
	ctx context.Context,
	client descAuth.AuthV1Client,
	refreshToken string,
) (*descAuth.GetAccessTokenResponse, error) {
	return client.GetAccessToken(
		ctx, &descAuth.GetAccessTokenRequest{
			RefreshToken: refreshToken,
		},
	)
}

func handleUnauthenticatedError[T any](
	ctx context.Context,
	err error,
	st *storage.Storage,
	res *T,
	baseErrMsg string,
	retryFn func(ctx context.Context) (*T, error),
) {
	var se GRPCStatusInterface
	if errors.As(err, &se) {
		errMessage := se.GRPCStatus().Message()
		if errMessage == messages.AccessTokenInvalid {
			refreshAccessToken(ctx, st)
			ctx = getRequestContext(ctx, st)
			res, err = retryFn(ctx)
			if err != nil {
				handleError(err, baseErrMsg)
			}
		} else {
			logger.ErrorWithExit("%s: %s", baseErrMsg, errMessage)
		}
	} else {
		logger.ErrorWithExit("%s: %s", baseErrMsg, err)
	}
}

func handleError(err error, baseErrMsg string) {
	var se GRPCStatusInterface
	if errors.As(err, &se) {
		logger.ErrorWithExit("%s: %s", baseErrMsg, se.GRPCStatus().Message())
	}

	logger.ErrorWithExit("%s: %s", baseErrMsg, err)
}
