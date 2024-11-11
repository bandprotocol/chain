package cmd

import (
	"fmt"

	rpcclient "github.com/cometbft/cometbft/rpc/client"
	"github.com/cometbft/cometbft/rpc/client/http"
)

func createClients(nodeURIs []string) ([]rpcclient.RemoteClient, func(), error) {
	clients := make([]rpcclient.RemoteClient, 0, len(nodeURIs))
	for _, uri := range nodeURIs {
		httpClient, err := http.New(uri, "/websocket")
		if err != nil {
			continue
		}

		if err = httpClient.Start(); err != nil {
			continue
		}

		clients = append(clients, httpClient)
	}

	if len(clients) == 0 {
		return nil, nil, fmt.Errorf("no clients are available")
	}

	// Function to stop all clients created so far
	stopClients := func() {
		for _, client := range clients {
			_ = client.Stop()
		}
	}

	return clients, stopClients, nil
}
