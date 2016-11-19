// Init:
//   Print ":- dynamic notification/4." to file
// Insert:
//   assert(notification(uuid, subject, body, false).
//   listing(notification/4).
// Delete:
//   retract(notification(uuid, _, _, _).
//   listing(notification/4).
// Update:
//   (up to caller: not implemented) combine Delete and Insert
// Get one:
//   notification(uuid, Subject, Body, Dismissed), !.
// Get all:
//   notification(UUID, Subject, Body, false).
//
// Note that all values are expected to be "STRINGS"
package main

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
)

type program string

type Predicate struct {
	name  string
	arity int
}

func ToPredicateString(head string, params []string) string {
	answer := head + "("
	for i, p := range params {
		if i == len(params)-1 {
			answer = answer + "\"" + p + "\")"
		} else {
			answer = answer + "\"" + p + "\", "
		}
	}
	return answer
}

func (p Predicate) String() string {
	return p.name + "/" + strconv.Itoa(p.arity)
}

type query struct {
	query     string
	variables []string
}

// A fully grounded predicate, one with no variables
// used in insert and delete
type Fact string

type queryResult map[string]string

type prologDb struct {
	file    string
	schema  []Predicate
	dbMutex *sync.Mutex
}

func createPrologDb(file string, schema []Predicate) *prologDb {
	return &prologDb{file, schema, new(sync.Mutex)}
}

// Creates the prolog database from the file. If the file doesn't exist it will be created.
// Changing schemas on an existing file is not supported
func (db *prologDb) Init() error {
	_, err := os.Stat(db.file)
	if err != nil && os.IsNotExist(err) {
		f, err := os.Create(db.file)
		if err != nil {
			return err
		}
		for _, p := range db.schema {
			f.WriteString(":- dynamic " + p.name + "/" + strconv.Itoa(p.arity) + ".\n")
		}
		f.Close()
	} else if err != nil {
		return err
	}
	return nil
}

func (db *prologDb) call(toplevel string) (io.ReadCloser, error) {
	cmd := exec.Command("swipl", "-l", db.file, "-q", "-g", "true", "-t", toplevel)
	out, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	log.Debugf("swi args: %v", cmd.Args)
	return out, cmd.Start()
}

func (db *prologDb) callAndReplace(toplevel string) error {
	var buffer bytes.Buffer
	out, err := db.call(toplevel)
	if err != nil {
		return err
	}
	_, err = buffer.ReadFrom(out)
	if err != nil {
		return err
	}

	file, err := os.Create(db.file)
	if err != nil {
		return err
	}
	defer file.Close()
	buffer.WriteTo(file)
	return err
}

func (db *prologDb) GetAll(q query) []queryResult {
	db.dbMutex.Lock()
	defer db.dbMutex.Unlock()
	return nil
}

func (db *prologDb) allListings() string {
	asStrings := make([]string, len(db.schema))
	for i, p := range db.schema {
		asStrings[i] = p.String()
	}
	return "[" + strings.Join(asStrings, ",") + "]"
}

func (db *prologDb) Insert(f Fact) error {
	db.dbMutex.Lock()
	defer db.dbMutex.Unlock()

	toplevel := "assert(" + string(f) + "), listing(" + db.allListings() + ")."
	return db.callAndReplace(toplevel)
}

func (db *prologDb) Delete(f Fact) error {
	db.dbMutex.Lock()
	defer db.dbMutex.Unlock()

	toplevel := "retract(" + string(f) + "), listing(" + db.allListings() + ")."
	return db.callAndReplace(toplevel)
}
