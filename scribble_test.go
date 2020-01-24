package scribble_test

import (
	"encoding/gob"
	"io/ioutil"
	"math/rand"
	"os"
	"testing"

	"github.com/lucacasonato/scribble"
)

type Fish struct {
	Type string
}

var (
	db         *scribble.Document
	database   = "./deep/school"
	collection = "fish"
	onefish    = Fish{}
	twofish    = Fish{}
	redfish    = Fish{Type: "red"}
	bluefish   = Fish{Type: "blue"}
)

func TestMain(m *testing.M) {

	// remove any thing for a potentially failed previous test
	os.RemoveAll("./deep")

	// run
	code := m.Run()

	// cleanup
	os.RemoveAll("./deep")

	// exit
	os.Exit(code)
}

// Tests creating a new database, and using an existing database
func TestNew(t *testing.T) {
	// database should not exist
	if _, err := os.Stat(database); err == nil {
		t.Error("Expected nothing, got database")
	}

	// create a new database
	createDB()

	// database should exist
	if _, err := os.Stat(database); err != nil {
		t.Error("Expected database, got nothing")
	}

	// should use existing database
	createDB()

	// database should exist
	if _, err := os.Stat(database); err != nil {
		t.Error("Expected database, got nothing")
	}
}

func TestWriteAndRead(t *testing.T) {

	createDB()

	// add fish to database
	if err := db.Collection(collection).Document("redfish").Write(redfish); err != nil {
		t.Error("Create fish failed: ", err.Error())
	}

	// read fish from database
	if err := db.Collection(collection).Document("redfish").Read(&onefish); err != nil {
		t.Error("Failed to read: ", err.Error())
	}

	//
	if onefish.Type != "red" {
		t.Error("Expected red fish, got: ", onefish.Type)
	}

	destroySchool()
}

func TestGetAllDocuments(t *testing.T) {

	createDB()
	createSchool()

	fish, err := db.Collection(collection).GetAllDocuments()
	if err != nil {
		t.Error("Failed to read: ", err.Error())
	}

	if len(fish) <= 0 {
		t.Error("Expected some fish, have none")
	}

	destroySchool()
}

func TestGetDocuments(t *testing.T) {

	createDB()
	createSchool()

	fish, err := db.Collection(collection).GetDocuments(1, 3)
	if err != nil {
		t.Error("Failed to read: ", err.Error())
	}

	if len(fish) <= 0 {
		t.Error("Expected some fish, have none")
	}

	destroySchool()
}

func TestWriteAndReadEmpty(t *testing.T) {

	createDB()

	// create a fish with no home
	if err := db.Collection("").Document("redfish").Write(redfish); err == nil {
		t.Error("Allowed write of empty resource")
	}

	// create a home with no fish
	if err := db.Collection(collection).Document("").Write(redfish); err == nil {
		t.Error("Allowed write of empty resource")
	}

	// no place to read
	if err := db.Collection("").Document("redfish").Read(onefish); err == nil {
		t.Error("Allowed read of empty resource")
	}

	destroySchool()
}

func TestDelete(t *testing.T) {

	createDB()

	// add fish to database
	if err := db.Collection(collection).Document("redfish").Write(redfish); err != nil {
		t.Error("Create fish failed: ", err.Error())
	}

	// delete the fish
	if err := db.Collection(collection).Document("redfish").Delete(); err != nil {
		t.Error("Failed to delete: ", err.Error())
	}

	// read fish from database
	if err := db.Collection(collection).Document("redfish").Read(&onefish); err == nil {
		t.Error("Expected nothing, got fish")
	}

	destroySchool()
}

func TestDeleteall(t *testing.T) {

	createDB()
	createSchool()

	if err := db.Collection(collection).Delete(); err != nil {
		t.Error("Failed to delete: ", err.Error())
	}

	fish, err := db.Collection(collection).GetAllDocuments()
	if err == nil {
		t.Error("Expected nothing, have fish:", err.Error())
	}

	if len(fish) > 0 {
		t.Error("Expected nothing, have fish")
	}

	destroySchool()
}

// create a new scribble database
func createDB() error {
	var err error
	if os.Getenv("SCRIBBLE_TEST_JSON") == "1" {
		if db, err = scribble.NewJSON(database); err != nil {
			return err
		}
	} else {
		if db, err = scribble.New(database); err != nil {
			return err
		}
	}

	return nil
}

// create a fish
func createFish(fish Fish) error {
	return db.Collection(collection).Document(fish.Type).Write(fish)
}

// create many fish
func createSchool() error {
	for _, f := range []Fish{{Type: "red"}, {Type: "blue"}} {
		if err := db.Collection(collection).Document(f.Type).Write(f); err != nil {
			return err
		}
	}

	return nil
}

// destroy a fish
func destroyFish(name string) error {
	return db.Collection(collection).Document(name).Delete()
}

// destroy all fish
func destroySchool() error {
	return db.Collection(collection).Delete()
}

func BenchmarkWrite(b *testing.B) {
	os.RemoveAll("./deep")

	createDB()

	data := generateData(b.N)
	collection := db.Collection("fish")

	b.ResetTimer()
	b.ReportAllocs()
	b.StartTimer()

	for x := range data {
		b.StopTimer()
		os.MkdirAll("./deep/fish/"+x, 0755)

		b.StartTimer()

		collection.Document(x).Write(data[x])
	}

	b.StopTimer()
	os.RemoveAll("./deep")
}

func BenchmarkRead(b *testing.B) {
	createDB()

	data := generateData(b.N)
	collection := db.Collection("fish")

	for x := range data {
		collection.Document(x).Write(data[x])
	}

	b.ResetTimer()
	b.ReportAllocs()
	b.StartTimer()

	var da interface{}

	for x := range data {
		collection.Document(x).Read(da)
	}

	b.StopTimer()
	os.RemoveAll("./deep")
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func generateData(n int) (data map[string]*Fish) {
	data = make(map[string]*Fish, n)

	for i := 0; i < n; i++ {
		data[randSeq(10)] = &Fish{
			Type: randSeq(128),
		}
	}

	return
}

func BenchmarkWriteNative(b *testing.B) {
	os.RemoveAll("./native")

	data := generateData(b.N)

	b.ResetTimer()
	b.ReportAllocs()
	b.StartTimer()

	for x, d := range data {
		ioutil.WriteFile("./native/"+x, []byte(d.Type+d.Type), 0666)
	}

	b.StopTimer()
	os.RemoveAll("./native")
}

func BenchmarkWriteNativeStream(b *testing.B) {
	os.RemoveAll("./native")

	data := generateData(b.N)

	b.ResetTimer()
	b.ReportAllocs()
	b.StartTimer()

	for x, d := range data {
		f, _ := os.Create("./native/" + x)

		gob.NewEncoder(f).Encode(d)

		f.Close()

		// move final file into place
	}

	b.StopTimer()
	os.RemoveAll("./native")
}

var g interface{}

func BenchmarkReadNative(b *testing.B) {
	os.RemoveAll("./native")

	data := generateData(b.N)

	for x, d := range data {
		ioutil.WriteFile("./native/"+x, []byte(d.Type+d.Type), 0666)
	}

	b.ResetTimer()
	b.ReportAllocs()
	b.StartTimer()

	for x := range data {
		p, _ := ioutil.ReadFile("./native/" + x)
		g = p

	}

	b.StopTimer()
	os.RemoveAll("./native")
}

func BenchmarkReadNativeStream(b *testing.B) {
	os.RemoveAll("./native")

	data := generateData(b.N)

	for x, d := range data {
		f, _ := os.Create("./native/" + x)

		gob.NewEncoder(f).Encode(d)

		f.Close()

		// move final file into place
	}

	b.ResetTimer()
	b.ReportAllocs()
	b.StartTimer()

	for x := range data {
		f, _ := os.Open("./native/" + x)

		var p interface{}
		gob.NewDecoder(f).Decode(p)

		g = p

		f.Close()

		// move final file into place
	}

	b.StopTimer()
	os.RemoveAll("./native")
}
