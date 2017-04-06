package logger

import (
	"os"

	"github.com/op/go-logging"
)

var logger = logging.MustGetLogger("pget")

func InitLogger(debug bool) {
	format := logging.MustStringFormatter(
		`%{color}%{time:06-01-02 15:04:05.000} %{level:.4s} @%{shortfile}%{color:reset} %{message}`,
	)
	logging.SetFormatter(format)
	backend := logging.NewLogBackend(os.Stdout, "", 0)
	logging.SetBackend(backend)
	if debug {
		logging.SetLevel(logging.DEBUG, "pget")
	} else {
		logging.SetLevel(logging.INFO, "pget")
	}
}

func GetLogger() *logging.Logger {
	return logger
}
