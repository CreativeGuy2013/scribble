Scribble (FireScribble Edition) [![GoDoc](https://godoc.org/github.com/boltdb/bolt?status.svg)](http://godoc.org/github.com/creativeguy2013/scribble) [![Go Report Card](https://goreportcard.com/badge/github.com/creativeguy2013/scribble)](https://goreportcard.com/report/github.com/creativeguy2013/scribble) [![Build Status](https://travis-ci.org/CreativeGuy2013/scribble.svg?branch=master)](https://travis-ci.org/CreativeGuy2013/scribble)
--------

A tiny GOB based database in Golang - behaviour is very similar to Google Cloud Firestore

**Note**
If you would rather use JSON instad of GOB please use a version prior to 3.0.0. You will have less functionality and a slower db but it will be human readable.

### Installation

Install using `go get github.com/creativeguy2013/scribble`.

### Usage

```go
// a new scribble document, providing the directory where it will be writing to
db, err := scribble.New(dir)
if err != nil {
  fmt.Println("Error", err)
}

// open a collection from the base document
fishCollection := db.Collection("fish")

// open the document we want to write to
onefishDocument := fishCollection.Document("onefish")

// write the data to the document
fish := Fish{}
if err := onefishDocument.Write(fish); err != nil {
  fmt.Println("Error", err)
}

// Read a data from the database
onefish := Fish{}
if err := onefishDocument.Read(&onefish); err != nil {
  fmt.Println("Error", err)
}

// Read all fish from the database, returning an array of documents.
records, err := fishCollection.GetAlDocuments()
if err != nil {
  fmt.Println("Error", err)
}

fishies := []Fish{}
for _, f := range records {
  fishFound := Fish{}
  if err := f.Read(&onefish); err != nil {
    fmt.Println("Error", err)
  }
  fishies = append(fishies, fishFound)
}

// Read a select view of the fish from the database, in this case everything from index 1 to 3
records, err = fishCollection.GetDocuments(1, 3)
if err != nil {
	fmt.Println("Error 5", err)
}
// records has length 2
fmt.Println(len(records))

fishies = []Fish{}
for _, f := range records {
	fishFound := Fish{}
	if err := f.Read(&fishFound); err != nil {
		fmt.Println("Error 6", err)
	}
	fishies = append(fishies, fishFound)
}
fmt.Println(fishies)

// Delete a fish from the database
if err := onefishDocument.Delete(); err != nil {
  fmt.Println("Error", err)
}

// Delete all fish from the database
if err := fishCollection.Delete(); err != nil {
  fmt.Println("Error", err)
}

// Make a subcollection in a document
fishBabiesCollection := onefishDocument.Collection("babies")

// Make a make a document in a collection
firstbabyDocument := Document("firstbaby")

```

It is also possible to store a subcollection and data in the same document:

```go
starFish := db.Collection("fish").Document("starFish")

starFish.Write(map[string]bool{
  "isAwesome": true,
})

starFish.Collection("properties").Document("arms").Write(6)
```


## Documentation
- Complete documentation is available on [godoc](http://godoc.org/github.com/creativeguy2013/scribble).
- Coverage Report is available on [gocover](https://gocover.io/github.com/creativeguy2013/scribble)

## Todo/Doing
- Support for windows
- More methods to allow different types of reads/writes
- More tests (you can never have enough!)
- loading part into memory/caching