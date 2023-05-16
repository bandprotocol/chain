package cylinder

// Workers represents a collection of Worker instances.
type Workers []Worker

// Worker defines the interface for a worker that can be started and stopped.
type Worker interface {
	Start()
	Stop()
}
