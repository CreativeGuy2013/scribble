// Package scribble is a tiny gob/json database
package scribble

import (
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"sync"
)

// Version is the current version of the project
const Version = "4.5.1"

type (
	//Collection a collection of documents
	Collection struct {
		ID  string
		dir string
		err error
		e   encoder
	}

	//Document a single document which can have sub collections
	Document struct {
		ID  string
		dir string
		err error
		e   encoder
	}
)

var (
	mutex       = &sync.Mutex{}
	fileMutexes = make(map[string]*sync.Mutex)
)

// New creates a new scribble database at the desired directory location, and
// returns a *Document to then use for interacting with the database
func New(dir string) (*Document, error) {
	return new(dir, gobEncoder{})
}

// NewJSON creates a new json scribble database at the desired directory location, and
// returns a *Document to then use for interacting with the database
func NewJSON(dir string) (*Document, error) {
	return new(dir, jsonEncoder{})
}

func new(dir string, e encoder) (*Document, error) {
	//Clean the filepath before using it
	dir = filepath.Clean(dir)

	document := Document{
		dir: dir,
		e:   e,
	}

	// if the collection doesn't exist create it
	if _, err := os.Stat(filepath.Join(document.dir, "doc."+document.e.extension())); err == nil {
		return &document, nil
	}

	if _, err := os.Stat(document.dir); err != nil {
		if err := os.MkdirAll(document.dir, 0755); err != nil {
			return nil, err
		}
	}

	// if the document doesn't exist create it
	return &document, ioutil.WriteFile(filepath.Join(document.dir, "doc."+document.e.extension()), []byte("{}"), 0644)
}

//Document gets a document from a collection
func (c *Collection) Document(key string) *Document {
	if is, err := c.Check(); is {
		return &Document{
			dir: "",
			err: fmt.Errorf("something has failed previously, use c.Check() to check for errors: %s", err.Error()),
			e:   c.e,
		}
	}
	if key == "" {
		return &Document{
			dir: "",
			err: fmt.Errorf("key for document is empty"),
			e:   c.e,
		}
	}

	dir := filepath.Join(c.dir, key)

	return &Document{
		ID:  key,
		dir: dir,
		e:   c.e,
	}
}

//Collection gets a collction from in a document
func (d *Document) Collection(name string) *Collection {
	if is, err := d.Check(); is {
		return &Collection{
			dir: "",
			err: fmt.Errorf("something has failed previously, use c.Check() to check for errors: %s", err.Error()),
			e:   d.e,
		}
	}

	if name == "" {
		return &Collection{
			dir: "",
			err: fmt.Errorf("name for collection is empty"),
			e:   d.e,
		}
	}

	dir := filepath.Join(d.dir, name)

	return &Collection{
		ID:  name,
		dir: dir,
		e:   d.e,
	}
}

// Write locks the database and attempts to write the record to the database under
// the [collection] specified with the [resource] name given
func (d *Document) Write(v interface{}) error {
	var err error
	var is bool

	// check if there was an error
	if is, err = d.Check(); is {
		return fmt.Errorf("something has failed previously, use c.Check() to check for errors: %s", err.Error())
	}

	// ensure there is a place to save record
	if d.dir == "" {
		return fmt.Errorf("missing document - no place to save record")
	}

	if err = os.MkdirAll(d.dir, 0755); err != nil {
		return err
	}

	mutex := getMutex(d.dir)
	mutex.Lock()
	defer mutex.Unlock()

	finalPath := filepath.Join(d.dir, "doc."+d.e.extension())
	tempPath := finalPath + ".tmp"

	b, err := os.Create(tempPath)
	defer b.Close()
	if err != nil {
		return err
	}

	err = d.e.encode(b, v)
	if err != nil {
		return err
	}

	// move final file into place
	err = os.Rename(tempPath, finalPath)
	if err != nil {
		return err
	}

	return nil
}

// Read a record from the database
func (d *Document) Read(v interface{}) error {
	var err error
	var is bool

	// check if there was an error
	if is, err = d.Check(); is {
		return fmt.Errorf("something has failed previously, use c.Check() to check for errors: %s", err.Error())
	}

	// ensure there is a place to save record
	if d.dir == "" {
		return fmt.Errorf("missing collection - no place to save record")
	}

	var b *os.File

	// read record from database
	b, err = os.Open(filepath.Join(d.dir, "doc."+d.e.extension()))
	defer b.Close()
	if err != nil {
		return err
	}

	// decode data
	err = d.e.decode(b, v)
	if err != nil {
		return err
	}

	return nil
}

// GetDocuments gets documents in a collection starting from start til end, if start
func getDocuments(dir string, start, end int, e encoder) ([]*Document, error) {
	// check to see if collection (directory) exists
	if file, err := os.Stat(dir); err != nil || !file.IsDir() {
		return nil, err
	}

	// check to see if collection (directory) exists
	if _, err := os.Stat(dir); err != nil {
		return nil, err
	}

	// read all the files in the transaction.Collection
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("could not open the documents: %s", err.Error())
	}

	if end != 0 {
		// end > len(files) will throw an runtime error
		if end > len(files) {
			end = len(files)
		}
		// make only include the files that are requested
		files = files[start:end]
	}

	records := make([]*Document, len(files))
	// iterate over each of the files, and add the resulting document to records
	for i, file := range files {
		records[i] = &Document{
			dir: filepath.Join(dir, file.Name()),
			e:   e,
		}
	}

	// unmarhsal the read files as a comma delimeted byte array
	return records, nil
}

// GetAllDocuments gets all documents in a collection.
func (c *Collection) GetAllDocuments() ([]*Document, error) {
	if is, err := c.Check(); is {
		return nil, fmt.Errorf("something has failed previously, use c.Check() to check for errors: %s", err.Error())
	}
	return getDocuments(c.dir, 0, 0, c.e)
}

// GetDocuments gets documents in a collection starting from start til end, if start
func (c *Collection) GetDocuments(start, end int) ([]*Document, error) {
	if is, err := c.Check(); is {
		return nil, fmt.Errorf("something has failed previously, use c.Check() to check for errors: %s", err.Error())
	}
	return getDocuments(c.dir, start, end, c.e)
}

func delete(dir string) error {
	mutex := getMutex(dir)
	mutex.Lock()
	defer mutex.Unlock()

	// if fi is nil or error is not nil return
	if _, err := os.Stat(dir); err != nil {
		return err
	}

	return os.RemoveAll(dir)
}

// Delete locks that database and removes the document including all of its sub documents
func (d *Document) Delete() error {
	// check if there was an error
	if is, err := d.Check(); is {
		return fmt.Errorf("something has failed previously, use c.Check() to check for errors: %s", err.Error())
	}

	return delete(d.dir)
}

// Delete removes a collection and all of its childeren
func (c *Collection) Delete() error {
	// check if there was an error
	if is, err := c.Check(); is {
		return fmt.Errorf("something has failed previously, use c.Check() to check for errors: %s", err.Error())
	}

	return delete(c.dir)
}

//Check if there is an error while getting the collection
func (c *Collection) Check() (bool, error) {
	if c.err != nil {
		return true, c.err
	}

	return false, nil
}

//Check if there is an error while getting the document
func (d *Document) Check() (bool, error) {
	if d.err != nil {
		return true, d.err
	}

	return false, nil
}

//PreGen does a check to see if there is an error while getting the collection and make them if they dont exist yet.
func (c *Collection) PreGen() (bool, error) {
	if c.err != nil {
		return true, c.err
	}

	_, err := os.Stat(c.dir)
	if os.IsNotExist(err) {
		os.MkdirAll(c.dir, 0755)
		return false, nil
	}

	if err != nil {
		return true, err
	}

	return false, nil
}

//PreGen does a check to see if there is an error while getting the documents and make them if they dont exist yet.
func (d *Document) PreGen() (bool, error) {
	if d.err != nil {
		return true, d.err
	}

	_, err := os.Stat(d.dir)
	if os.IsNotExist(err) {
		b, err := os.Create(filepath.Join(d.dir, "doc."+d.e.extension()))
		defer b.Close()
		if err != nil {
			return true, err
		}
		return false, nil
	}

	if err != nil {
		return true, err
	}

	return false, nil
}

// getMutex gets a mutex for a specific dir
func getMutex(dir string) *sync.Mutex {

	mutex.Lock()
	defer mutex.Unlock()
	m, ok := fileMutexes[dir]

	// if the mutex doesn't exist make it
	if !ok {
		fileMutexes[dir] = &sync.Mutex{}
		return fileMutexes[dir]
	}
	return m
}

type encoder interface {
	extension() string
	encode(io.Writer, interface{}) error
	decode(io.Reader, interface{}) error
}

type gobEncoder struct{}

func (e gobEncoder) extension() string {
	return "gob"
}
func (e gobEncoder) encode(b io.Writer, v interface{}) error {
	return gob.NewEncoder(b).Encode(v)
}
func (e gobEncoder) decode(b io.Reader, v interface{}) error {
	if rv, ok := v.(reflect.Value); ok {
		return gob.NewDecoder(b).DecodeValue(rv)
	}
	return gob.NewDecoder(b).Decode(v)
}

type jsonEncoder struct{}

func (e jsonEncoder) extension() string {
	return "json"
}
func (e jsonEncoder) encode(b io.Writer, v interface{}) error {
	return json.NewEncoder(b).Encode(v)
}
func (e jsonEncoder) decode(b io.Reader, v interface{}) error {
	return json.NewDecoder(b).Decode(v)
}
