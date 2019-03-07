package main

import (
	"fmt"

	scribble "github.com/creativeguy2013/scribble"
)

//Fish is a fish struct
type Fish struct{ Name string }

func main() {

	dir := "./db"

	db, err := scribble.New(dir)
	if err != nil {
		fmt.Println("Error 1", err)
	}

	fishCollection := db.Collection("fish")

	// Write a fish to the database
	for _, name := range []string{"onefish", "twofish", "redfish", "bluefish"} {
		if err := fishCollection.Document(name).Write(Fish{Name: name}); err != nil {
			fmt.Println("Error 2", err)
		}
	}

	fishCollection.Document("purple").Write(Fish{
		Name: "purple",
	})

	fishCollection.Document("purple").Collection("teeth").Document("3").Write(map[string]string{
		"one": "two",
		"two": "three",
	})

	// Read a fish from the database (passing fish by reference)
	onefish := Fish{}
	if err := fishCollection.Document("onefish").Read(&onefish); err != nil {
		fmt.Println("Error 3", err)
	}

	// Read all fish from the database, unmarshaling the response.
	records, err := fishCollection.GetAllDocuments()
	if err != nil {
		fmt.Println("Error 4", err)
	}
	// records has length 5
	fmt.Println(len(records))

	fishies := []Fish{}
	for _, f := range records {
		fishFound := Fish{}
		if err := f.Read(&fishFound); err != nil {
			fmt.Println("Error 6", err)
		}
		fishies = append(fishies, fishFound)
	}
	fmt.Println(fishies)

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
}
