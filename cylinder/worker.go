package cylinder

type Workers []Worker

type Worker interface {
	Start()
	Stop()
}
