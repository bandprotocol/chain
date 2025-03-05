package cylinder

import "github.com/bandprotocol/chain/v3/cylinder/context"

// Run starts the Cylinder process with the provided context and workers.
func Run(c *context.Context, workers Workers) error {
	c.Logger.Info(":star: Start each worker:")

	// Start all workers concurrently
	for _, worker := range workers {
		go worker.Start()
	}

	err := <-c.ErrCh

	// Stop all workers if there was an error
	for _, worker := range workers {
		if err := worker.Stop(); err != nil {
			return err
		}
	}

	return err
}
