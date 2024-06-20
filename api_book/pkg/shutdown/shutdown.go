package shutdown

import (
	"io"
	"os"
	"os/signal"
)

type logger interface {
	Infof(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

func Graceful(logger logger, signals []os.Signal, closeItems ...io.Closer) {
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, signals...)
	sig := <-sigc
	logger.Infof("caught signal %s. Shutting down...", sig)

	for _, closer := range closeItems {
		if err := closer.Close(); err != nil {
			logger.Errorf("failed to close %v: %v", closer, err)
		}
	}
}
