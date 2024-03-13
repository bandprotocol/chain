package priceservice

import (
	"context"
	"time"

	bothanproto "github.com/bandprotocol/bothan-api/go-proxy/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GRPCService struct {
	url     string
	timeout time.Duration
}

func NewGRPCService(url string, timeout time.Duration) *GRPCService {
	return &GRPCService{url: url, timeout: timeout}
}

func (gs *GRPCService) Query(signalIds []string) ([]*bothanproto.PriceData, error) {
	// Set up a connection to the server.
	conn, err := grpc.Dial(gs.url, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// Create a client instance using the connection.
	client := bothanproto.NewQueryClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), gs.timeout)
	defer cancel()

	response, err := client.Prices(ctx, &bothanproto.QueryPricesRequest{SignalIds: signalIds})
	if err != nil {
		return nil, err
	}

	return response.Prices, nil
}
