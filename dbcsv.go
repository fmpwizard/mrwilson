package main

import (
	"encoding/csv"
	"io"
	"log"
	"net/http"
	"os"
)

//CSVHandler gives you info about the db
func CSVHandler(w http.ResponseWriter, r *http.Request) {

	f, err := os.OpenFile(db, os.O_RDONLY|os.O_CREATE, 0600)
	if err != nil {
		log.Fatalf("error opening file %s, got: %s", db, err)
		return
	}
	csvReader := csv.NewReader(f)
	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		for _, cell := range record {
			w.Write([]byte(cell))
		}
		w.Write([]byte("\n"))
	}
}
