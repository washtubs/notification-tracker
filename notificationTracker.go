package main

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

func genUUID() string {
	//TODO
	return "1"
}

func (db *notificationDb) add(subject string, body string) error {
	f := Fact(ToPredicateString("notification", []string{genUUID(), subject, body, "false"}))
	return db.delegate.Insert(f)
}
