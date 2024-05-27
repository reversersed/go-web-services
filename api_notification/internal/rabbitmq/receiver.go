package rabbitmq

type Receiver interface {
	Close() error
	Start()
}
