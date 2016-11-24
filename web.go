package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

const FORBIDDEN int = 403

var forbiddenResp http.Response = http.Response{
	StatusCode: FORBIDDEN,
}

var errorResp http.Response = http.Response{
	StatusCode: 500,
}

func handleNew(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		forbiddenResp.Write(w)
		return
	}
	subject := req.FormValue("subject")
	body := req.FormValue("body")

	id, err := stdNotificationDb.add(subject, body, "", false)
	if err != nil {
		log.Error(err)
		errorResp.Write(w)
		return
	}

	w.Write([]byte(id))
}

func handleDismiss(w http.ResponseWriter, req *http.Request) {
	if req.Method != "PUT" {
		forbiddenResp.Write(w)
		return
	}

	url := req.URL
	stringId := strings.TrimPrefix(url.EscapedPath(), "/dismiss/")
	// TODO: validate stringId
	stdNotificationDb.markDismissed(stringId, true)
}

func handleList(w http.ResponseWriter, req *http.Request) {
	if req.Method != "GET" {
		forbiddenResp.Write(w)
		return
	}

	//marshalled, err := json.Marshal(&Notification{"123", "sss", "bbb", false})
	notifications, err := stdNotificationDb.listNonDismissed()
	if err != nil {
		log.Error(err)
		errorResp.Write(w)
		return
	}

	marshalled, err := json.Marshal(notifications)
	if err != nil {
		log.Error(err)
		errorResp.Write(w)
		return
	}

	w.Write(marshalled)
}
