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
	"bufio"
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

func (p Predicate) String() string {
	return p.name + "/" + strconv.Itoa(p.arity)
}

type Query struct {
	head      string
	params    []string
	variables []string
}

func (q *Query) withHead(s string) *Query {
	q.head = s
	return q
}

func (q *Query) underscoreParam() *Query {
	q.params = append(q.params, "_")
	return q
}

func (q *Query) stringParam(s string) *Query {
	q.params = append(q.params, "\""+s+"\"")
	return q
}

func (q *Query) varParam(s string) *Query {
	q.params = append(q.params, s)
	q.variables = append(q.variables, s)
	return q
}

func (q *Query) String() string {
	return q.head + "(" + strings.Join(q.params, ",") + ")"
}

// A fully grounded predicate, one with no variables
// used in insert and delete
type Fact string

type QueryResult map[string]string

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
	log.Debug("writing to file")
	buffer.WriteTo(file)
	return err
}

const DELIM = "|DELIM|"

func makeFormatString(size int) string {
	var fstring string
	for i := 0; i < size-1; i++ {
		fstring = fstring + "~w" + DELIM
	}
	fstring = fstring + "~w~n"
	return fstring
}

func (db *prologDb) GetAll(q *Query) ([]QueryResult, error) {
	db.dbMutex.Lock()
	defer db.dbMutex.Unlock()
	makeFormatString(len(q.variables))
	toplevel := "forall(" + q.String() + ", format('" + makeFormatString(len(q.variables)) + "', [" + strings.Join(q.variables, ",") + "]))"
	out, err := db.call(toplevel)
	if err != nil {
		return nil, err
	}
	sc := bufio.NewScanner(bufio.NewReader(out))
	results := make([]QueryResult, 0)
	for sc.Scan() {
		r := make(QueryResult)
		for i, value := range strings.Split(sc.Text(), DELIM) {
			r[q.variables[i]] = value
		}
		results = append(results, r)
	}

	return results, nil
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
