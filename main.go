package main

import (
	"flag"
	"net/http"
	"os"
	"strconv"

	logging "github.com/op/go-logging"
)

var log *logging.Logger
var logBackend *logging.LogBackend
var LeveledLogBackend logging.LeveledBackend

var stdConfig defaultConfig
var stdNotificationDb *notificationDb

func main() {
	log.Debugf("Starting notification-tracker...")
	var (
		port = flag.Int("port", 8133, "port number for the HTTP server")
		db   = flag.String("database", "", "Path to prolog database file")
	)
	flag.Parse()
	stdConfig = createNodeConfig(*port, *db)

	mux := http.NewServeMux()
	//mux.HandleFunc("/upload", handleUpload)

	server := http.Server{
		Addr:    ":" + strconv.Itoa(stdConfig.port),
		Handler: mux,
	}

	log.Infof("Creating notification database from file: %s", stdConfig.db)
	stdNotificationDb = createNotificationDatabase(stdConfig.db)
	err := stdNotificationDb.init()
	if err != nil {
		log.Fatal(err)
	}

	err = stdNotificationDb.add("subjecctt", "booooty")
	if err != nil {
		log.Fatal(err)
	}

	err = server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}

type defaultConfig struct {
	port int
	db   string
}

func createNodeConfig(port int, db string) defaultConfig {
	return defaultConfig{
		port,
		db,
	}
}

func init() {
	log = logging.MustGetLogger("notification-tracker")
	logBackend = logging.NewLogBackend(os.Stderr, "", 0)
	LeveledLogBackend = logging.AddModuleLevel(logBackend)
	LeveledLogBackend.SetLevel(logging.DEBUG, "")
	logging.SetBackend(LeveledLogBackend)

	logging.SetFormatter(logging.MustStringFormatter(
		`%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`,
	))
}
