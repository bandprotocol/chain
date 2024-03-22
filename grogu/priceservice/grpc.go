package priceservice

import (
	"context"
	"time"

	bothanproto "github.com/bandprotocol/bothan-api/go-proxy/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GRPCService struct {
	connection *grpc.ClientConn
	timeout    time.Duration
}

func NewGRPCService(url string, timeout time.Duration) (*GRPCService, error) {
	conn, err := grpc.Dial(url, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &GRPCService{connection: conn, timeout: timeout}, nil
}

func (gs *GRPCService) Query(signalIds []string) ([]*bothanproto.PriceData, error) {
	// Create a client instance using the connection.
	client := bothanproto.NewQueryClient(gs.connection)
	ctx, cancel := context.WithTimeout(context.Background(), gs.timeout)
	defer cancel()

	response, err := client.Prices(ctx, &bothanproto.QueryPricesRequest{SignalIds: signalIds})
	if err != nil {
		return nil, err
	}

	return response.Prices, nil
}
