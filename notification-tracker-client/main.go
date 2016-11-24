package main

import (
	"flag"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"

	logging "github.com/op/go-logging"
)

func userErr(message string) {
	panic(message)
}

var log *logging.Logger
var logBackend *logging.LogBackend
var LeveledLogBackend logging.LeveledBackend
var debugMode *bool

func main() {

	if len(os.Args) < 2 {
		userErr("Need an action argument")
	}

	host := os.ExpandEnv("$NOTIFICATION_TRACKER_HOST")

	if host == "" {
		userErr("Please set the environment variable NOTIFICATION_TRACKER_HOST.")
	}

	var flagSet *flag.FlagSet
	var execute func()

	//globalFlagSet.Arg(os.Args[2:])
	switch os.Args[1] {
	case "list":
		flagSet = flag.NewFlagSet("list", flag.ExitOnError)
		var (
			dismissed = flagSet.Bool("dismissed", false,
				"If you want to include notifications that have been dismissed")
		)
		execute = func() {
			list(host, *dismissed)
		}
	case "send":
		flagSet = flag.NewFlagSet("send", flag.ExitOnError)
		var (
			subject = flagSet.String("subject", "", "The subject of the notification")
			body    = flagSet.String("body", "", "The body of the notification")
		)
		execute = func() {
			if *subject == "" {
				log.Error("subject empty")
				return
			}
			if *body == "" {
				log.Error("body empty")
				return
			}
			send(host, *subject, *body)
		}
	default:
		userErr("Unrecognized action: " + os.Args[1])
	}

	debugMode = flagSet.Bool("v", false, "Verbose")

	flagSet.Parse(os.Args[2:])

	if *debugMode {
		LeveledLogBackend.SetLevel(logging.DEBUG, "")
	}

	execute()
}

func send(host string, subject string, body string) {
	resp, err := http.PostForm(host+"/new", url.Values{
		"subject": {subject},
		"body":    {body},
	})

	if err != nil {
		panic(err)
	}

	if resp.StatusCode != 200 {
		panic("Status code: " + strconv.Itoa(resp.StatusCode) + " " + resp.Status)
	}

	defer resp.Body.Close()
	io.Copy(os.Stdout, resp.Body)

}

func list(host string, dismissed bool) {
	if dismissed {
		panic("Including dismissed notifications is not supported yet!")
	}
	resp, err := http.Get(host + "/list")
	if err != nil {
		panic(err)
	}
	if resp.StatusCode != 200 {
		panic("Status code: " + strconv.Itoa(resp.StatusCode) + " " + resp.Status)
	}
	defer resp.Body.Close()
	io.Copy(os.Stdout, resp.Body)
}

func init() {
	log = logging.MustGetLogger("notification-client")
	logBackend = logging.NewLogBackend(os.Stderr, "", 0)
	LeveledLogBackend = logging.AddModuleLevel(logBackend)
	LeveledLogBackend.SetLevel(logging.DEBUG, "")
	logging.SetBackend(LeveledLogBackend)

	logging.SetFormatter(logging.MustStringFormatter(
		`%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`,
	))
}
