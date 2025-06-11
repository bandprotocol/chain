package cylinder

import "github.com/bandprotocol/chain/v3/cylinder/msg"

// Workers represents a collection of Worker instances.
type Workers []Worker

// Worker defines the interface for a worker that can be started and stopped.
type Worker interface {
	Start()
	Stop() error
	GetResponseReceivers() []*msg.ResponseReceiver
}
