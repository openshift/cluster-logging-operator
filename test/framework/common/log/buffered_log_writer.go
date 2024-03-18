package log

import (
	logger "github.com/ViaQ/logerr/v2/log"
	"github.com/go-logr/logr"
	"io"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sync"
)

func init() {
	log.SetLogger(zap.New())
}

// BufferedLogWriter intercepts log messages to allow control of its internal buffer is flushed
type BufferedLogWriter struct {
	mtx    sync.Mutex
	log    logr.Logger
	out    io.Writer
	buffer []byte
}

func NewBufferedLogWriter() *BufferedLogWriter {
	return &BufferedLogWriter{
		log:    logger.NewLogger("internal"),
		out:    os.Stdout,
		buffer: []byte{},
	}
}

// Flush the contents of the sink
func (w *BufferedLogWriter) Flush() {
	w.mtx.Lock()
	defer w.mtx.Unlock()
	if _, err := w.out.Write(w.buffer); err != nil {
		w.log.Error(err, "error flushing log messages")
	}
	w.buffer = []byte{}
}

func (w *BufferedLogWriter) Write(p []byte) (n int, err error) {
	w.mtx.Lock()
	defer w.mtx.Unlock()
	w.buffer = append(w.buffer, p...)
	return len(p), nil
}

func NewLogger(component string, verbosity int) (logr.Logger, *BufferedLogWriter) {
	writer := NewBufferedLogWriter()
	return logger.NewLogger(component,
		logger.WithVerbosity(verbosity),
		logger.WithOutput(writer)), writer

}
