package main

import (
	"errors"
	"math/rand"
	"strconv"
	"time"
)

// TODO: validations: subject and body cant be "_" as it's specially recognized
// subject and body bust fit on a line
type notificationDb struct {
	delegate *prologDb
}

type Notification struct {
	Id        string `json:"id"`
	Subject   string `json:"subject"`
	Body      string `json:"body"`
	Dismissed bool   `json:"dismissed"`
}

func notificationFromResult(result QueryResult, id string) Notification {
	return Notification{
		id,
		result["Subject"],
		result["Body"],
		strToBool(result["Dismissed"]),
	}
}
func notificationFromResultNonDismissed(result QueryResult) Notification {
	return Notification{
		result["Id"],
		result["Subject"],
		result["Body"],
		false,
	}
}

func createNotificationDatabase(file string) *notificationDb {
	notificationPred := Predicate{"notification", 4}
	prologDb := createPrologDb(file, []Predicate{notificationPred})
	return &notificationDb{prologDb}
}

func (db *notificationDb) init() error {
	return db.delegate.Init()
}

// http://stackoverflow.com/a/22892986
var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func genUUID() string {
	return randSeq(8)
}

func (db *notificationDb) add(subject string, body string, id string, dismissed bool) (string, error) {
	if id == "" {
		id = genUUID()
	}
	query := new(Query)
	query.
		withHead("notification").
		stringParam(id).
		stringParam(subject).
		stringParam(body).
		stringParam(boolToStr(dismissed))

	f := Fact(query.String())
	return id, db.delegate.Insert(f)
}

func boolToStr(b bool) string {
	if b {
		return "true"
	} else {
		return "false"
	}
}

func strToBool(s string) bool {
	if s == "true" {
		return true
	} else {
		return false
	}
}

func (db *notificationDb) getById(id string) (Notification, error) {
	//db.delegate.GetAll(q)
	query := new(Query)
	query.
		withHead("notification").
		stringParam(id).
		varParam("Subject").
		varParam("Body").
		varParam("Dismissed")
	results, err := db.delegate.GetAll(query)
	if err != nil {
		return Notification{}, err
	}
	if len(results) != 1 {
		return Notification{}, errors.New("Expected exactly one result: got " + strconv.Itoa(len(results)))
	} else {
		return notificationFromResult(results[0], id), nil
	}
}

func (db *notificationDb) listNonDismissed() ([]Notification, error) {
	query := new(Query)
	query.
		withHead("notification").
		varParam("Id").
		varParam("Subject").
		varParam("Body").
		stringParam(boolToStr(false))
	results, err := db.delegate.GetAll(query)
	if err != nil {
		return nil, err
	}
	notifications := make([]Notification, 0)
	for _, result := range results {
		notif := notificationFromResultNonDismissed(result)
		notifications = append(notifications, notif)
	}
	return notifications, nil
}

func (db *notificationDb) markDismissed(id string, dismissed bool) error {
	notif, err := db.getById(id)
	if err != nil {
		return err
	}

	query := new(Query)
	query.
		withHead("notification").
		stringParam(id).
		underscoreParam().
		underscoreParam().
		underscoreParam()

	err = db.delegate.Delete(Fact(query.String()))
	if err != nil {
		return err
	}

	_, err = db.add(notif.Subject, notif.Body, id, dismissed)
	if err != nil {
		return err
	}
	//db.delegate.Delete(Fact(ToPredicateString("notification", []string{id, "_", "_", boolToStr(dismissed)})))
	//f := Fact(ToPredicateString("notification", []string{genUUID(), subject, body, "false"}))
	//return id, db.delegate.Insert(f)
	return nil
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
