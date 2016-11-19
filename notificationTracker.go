package main

import (
	"math/rand"
)

// TODO: validations: subject and body cant be "_" as it's specially recognized
type notificationDb struct {
	delegate *prologDb
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

func (db *notificationDb) add(subject string, body string) (string, error) {
	id := genUUID()
	f := Fact(ToPredicateString("notification", []string{id, subject, body, "false"}))
	return id, db.delegate.Insert(f)
}

func boolToStr(b bool) string {
	if b {
		return "true"
	} else {
		return "false"
	}
}

func (db *notificationDb) getById(id string) (string, error) {
	return "", nil
}

func (db *notificationDb) markDismissed(id string, dismissed bool) (string, error) {
	//db.delegate.Delete(Fact(ToPredicateString("notification", []string{id, "_", "_", boolToStr(dismissed)})))
	//f := Fact(ToPredicateString("notification", []string{genUUID(), subject, body, "false"}))
	//return id, db.delegate.Insert(f)
	return "", nil
}
