// Package scribble is a tiny gob database
package scribble

import (
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
)

// Version is the current version of the project
const Version = "4.0.0"

type (
	//Collection a collection of documents
	Collection struct {
		dir string // the directory where scribble will create the database
		err error
	}

	//Document a single document which can have sub collections
	Document struct {
		dir string
		err error
	}
)

var (
	mutex       = &sync.Mutex{}
	fileMutexes = make(map[string]*sync.Mutex)
)

// New creates a new scribble database at the desired directory location, and
// returns a *Driver to then use for interacting with the database
func New(dir string) (*Document, error) {
	//Clean the filepath before using it
	dir = filepath.Clean(dir)

	document := Document{
		dir: dir,
	}

	// if the collection doesn't exist create it
	if _, err := os.Stat(filepath.Join(document.dir, "doc.gob")); err == nil {
		return &document, nil
	}

	if _, err := os.Stat(document.dir); err != nil {
		if err := os.MkdirAll(document.dir, 0755); err != nil {
			fmt.Println(err.Error())
			return nil, err
		}
	}

	// if the document doesn't exist create it
	return &document, ioutil.WriteFile(filepath.Join(document.dir, "doc.gob"), []byte("{}"), 0644)
}

//Document gets a document from a collection
func (c *Collection) Document(key string) *Document {
	if c.Error() != "" {
		return &Document{
			dir: "",
			err: fmt.Errorf("sometething has failled previously, use c.Error to check for errors: %s", c.err.Error()),
		}
	}
	if key == "" {
		return &Document{
			dir: "",
			err: fmt.Errorf("key for document is empty"),
		}
	}

	dir := filepath.Join(c.dir, key)

	document := Document{
		dir: dir,
	}

	return &document
}

//Collection gets a collction from in a document
func (d *Document) Collection(name string) *Collection {
	if err := d.Error(); err != "" {
		return &Collection{
			dir: "",
			err: fmt.Errorf("sometething has failled previously, use c.Error to check for errors: %s", d.err.Error()),
		}
	}

	if name == "" {
		return &Collection{
			dir: "",
			err: fmt.Errorf("name for collection is empty"),
		}
	}

	dir := filepath.Join(d.dir, name)

	collection := Collection{
		dir: dir,
	}

	return &collection
}

// Write locks the database and attempts to write the record to the database under
// the [collection] specified with the [resource] name given
func (d *Document) Write(v interface{}) error {
	// check if there was an error
	if error := d.Error(); error != "" {
		return fmt.Errorf("sometething has failled previously, use c.Error to check for errors: %s", error)
	}

	// ensure there is a place to save record
	if d.dir == "" {
		return fmt.Errorf("missing document - no place to save record")
	}

	if _, err := os.Stat(d.dir); err != nil {
		if err := os.MkdirAll(d.dir, 0755); err != nil {
			return err
		}
	}

	mutex := getMutex(d.dir)
	mutex.Lock()
	defer mutex.Unlock()

	dir := d.dir
	fnlPath := filepath.Join(dir, "doc.gob")
	tmpPath := fnlPath + ".tmp"

	// create collection directory
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	b, err := os.Create(tmpPath)
	if err != nil {
		return err
	}

	err = gob.NewEncoder(b).Encode(v)
	if err != nil {
		return err
	}

	// move final file into place
	return os.Rename(tmpPath, fnlPath)
}

// Read a record from the database
func (d *Document) Read(v interface{}) error {
	// check if there was an error
	if error := d.Error(); error != "" {
		return fmt.Errorf("sometething has failled previously, use c.Error to check for errors: %v", error)
	}

	// ensure there is a place to save record
	if d.dir == "" {
		return fmt.Errorf("missing collection - no place to save record")
	}

	//
	record := filepath.Join(d.dir, "doc.gob")

	// check to see if file exists
	if _, err := os.Stat(record); err != nil {
		return err
	}

	// read record from database
	b, err := os.Open(record)
	if err != nil {
		return err
	}

	// unmarshal data
	return gob.NewDecoder(b).Decode(v)
}

// GetDocuments gets documents in a collection starting from start til end, if start
func getDocuments(dir string, start, end int) ([]*Document, error) {
	// check to see if collection (directory) exists
	if file, err := os.Stat(dir); err != nil || !file.IsDir() {
		return nil, err
	}

  // check to see if collection (directory) exists
	if _, err := os.Stat(dir); err != nil {
		fmt.Println("2: ", err)
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
		}
	}

	// unmarhsal the read files as a comma delimeted byte array
	return records, nil
}

// GetAllDocuments gets all documents in a collection.
func (c *Collection) GetAllDocuments() ([]*Document, error) {
	if error := c.Error(); error != "" {
		return nil, fmt.Errorf("sometething has failled previously, use c.Error to check for errors: %v", error)
	}
	return getDocuments(c.dir, 0, 0)
}

// GetDocuments gets documents in a collection starting from start til end, if start
func (c *Collection) GetDocuments(start, end int) ([]*Document, error) {
	if error := c.Error(); error != "" {
		return nil, fmt.Errorf("sometething has failled previously, use c.Error to check for errors: %v", error)
	}
	return getDocuments(c.dir, start, end)
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
	if error := d.Error(); error != "" {
		return fmt.Errorf("sometething has failled previously, use c.Error to check for errors: %v", error)
	}

	return delete(d.dir)
}

// Delete removes a collection and all of its childeren
func (c *Collection) Delete() error {
	// check if there was an error
	if error := c.Error(); error != "" {
		return fmt.Errorf("sometething has failled previously, use c.Error to check for errors: %v", error)
	}

	return delete(c.dir)
}

//Error if there is an error while getting the collection
func (c *Collection) Error() string {
	if c.err != nil {
		return c.err.Error()
	}
	
  return ""
}

//Error if there is an error while getting the document
func (d *Document) Error() string {
	if d.err != nil {
		return d.err.Error()
	}
	return ""
}

func statOld(path string) (fi os.FileInfo, err error) {
	// check for dir, if path isn't a directory check to see if it's a file
	if fi, err = os.Stat(path); os.IsNotExist(err) {
		fi, err = os.Stat(filepath.Join(path, "doc.gob"))
	}

	return
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
