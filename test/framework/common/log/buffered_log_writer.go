package log

import (
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"sync"

	logger "github.com/ViaQ/logerr/v2/log"
	"github.com/go-logr/logr"
	"github.com/openshift/cluster-logging-operator/test/helpers/errors"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
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
	w := &BufferedLogWriter{
		log:    logger.NewLogger("internal"),
		out:    os.Stdout,
		buffer: []byte{},
	}
	return w
}

func (w *BufferedLogWriter) FlushToArtifactsDir(name string) {
	w.mtx.Lock()
	defer w.mtx.Unlock()
	if dir := os.Getenv("ARTIFACT_DIR"); dir != "" {
		if parts := strings.SplitAfter(name, "/test/"); len(parts) == 2 {
			name = strings.ReplaceAll(parts[1], "/", "_")
			fullPath := path.Join(dir, name)
			if file, err := os.Create(fullPath); err != nil {
				w.log.Error(err, fmt.Sprintf("Unable to flush logs to file: %s", fullPath))
			} else {
				w.out = file
				w.flushLocked()
				errors.LogIfError(file.Close())
				w.out = os.Stdout
				return
			}
		}
	}
	w.flushLocked()
}

func (w *BufferedLogWriter) flushLocked() {
	if _, err := w.out.Write(w.buffer); err != nil {
		w.log.Error(err, "error flushing log messages")
	}
	w.buffer = []byte{}
}

// Flush the contents of the sink
func (w *BufferedLogWriter) Flush() {
	w.mtx.Lock()
	defer w.mtx.Unlock()
	w.flushLocked()
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
