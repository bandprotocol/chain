package cylinder

// Run starts the Cylinder process with the provided context and workers.
func Run(c *Context, workers Workers) error {
	c.Logger.Info(":star: Start each worker:")

	// Start all workers concurrently
	for _, worker := range workers {
		go worker.Start()
	}

	err := <-c.ErrCh

	// Stop all workers if there was an error
	for _, worker := range workers {
		worker.Stop()
	}

	return err
}
