package gnuplot

import (
	logerr "github.com/ViaQ/logerr/v2/log"
	"github.com/go-logr/logr"
)

var (
	log logr.Logger
)

func init() {
	log = logerr.NewLogger("stats")
}
