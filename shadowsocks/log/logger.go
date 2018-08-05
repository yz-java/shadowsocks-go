package log

import (
	"os"
	"github.com/op/go-logging"
)
var Logger = logging.MustGetLogger("logger")

var format = logging.MustStringFormatter(
	`%{color}%{time:2006-01-02 15:04:05.000} %{shortfile} %{longfunc} >>> %{level:.4s} %{id:04d} %{message}%{color:reset}`,
)
func CreateLog(path string) {
	logFile, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0777)
	if err != nil {
		Logger.Fatal("open Logger file error", err)
	}
	backend1 := logging.NewLogBackend(logFile, "", 0)
	backend2 := logging.NewLogBackend(os.Stderr, "", 0)

	backend1Formatter := logging.NewBackendFormatter(backend1, format)
	backend2Formatter := logging.NewBackendFormatter(backend2, format)

	logging.SetBackend(backend1Formatter, backend2Formatter)
}




