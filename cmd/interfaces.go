package cmd

import "google.golang.org/grpc/status"

type GRPCStatusInterface interface {
	GRPCStatus() *status.Status
}
