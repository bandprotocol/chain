package cylinder

func Run(c *Context, workers Workers) error {
	c.Logger.Info(":star: Start each worker:")

	// start all workers
	for _, worker := range workers {
		go worker.Start()
	}

	err := <-c.ErrCh

	// stop all worker if something error
	for _, worker := range workers {
		worker.Stop()
	}

	return err
}
