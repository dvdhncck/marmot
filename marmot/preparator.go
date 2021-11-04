package marmot

import (
	"bufio"
	"database/sql"
	"fmt"

	//	"encoding/json"
	"path"
	"path/filepath"

	//	"github.com/disintegration/imaging"
	"io/ioutil"
	"log"
	"os"
	"strings"
	//	"path/filepath"
)


func Prepare(db *sql.DB, preparePath string) {
	// directory should be readable
	_, err := ioutil.ReadDir(preparePath)
	if err != nil {
		log.Fatal(err)
	}

	// check that we are pointing to something sensible
	fails := NewFails(preparePath)
	validateName(fails, preparePath)

	if !fails.IsGood() {
		fails.Write()
		return
	}

	metaFilePath := filepath.Join(preparePath, "meta.json")

	// guard against meta.json already there
	if _, err := os.Stat(metaFilePath); os.IsNotExist(err) {

		// allocate a new ID
		maxId := -1
		row := db.QueryRow("SELECT max(ID) FROM Album")
		err = row.Scan(&maxId)
		if err != nil {			
			log.Fatal(err.Error())
		}

		if maxId == -1 {
			log.Fatal(`Unable to get max ID`)
		}

		newId := maxId + 1

		log.Printf("Using ID %d", newId)

		// we've already validated the name, so we can proceed with incaution

		containerPath := path.Base(preparePath)
		parts := strings.Split(containerPath, "__")

		artist := enwordify(parts[0])
		title := enwordify(parts[1])

		handle, err := os.Create(metaFilePath)

		if err == nil {
			w := bufio.NewWriter(handle)
			json := fmt.Sprintf("{\n  \"id\":%d,\n  \"title\":\"%s\",\n  \"artists\":[\"%s\"],\n  \"genres\":[\"\"]\n}\n", newId, title, artist)
			_, err = w.WriteString(json)
			if err == nil {
				log.Printf("meta.json written to %s", metaFilePath)
			} else {
				log.Fatal(err)
			}
			w.Flush()
		} else {
			log.Fatal(err)
		}
	} else {
		log.Fatal(`meta.json already present in ` + metaFilePath)
	}

}

func enwordify(thing string) string {
	tokens := strings.Split(thing, "_")
	// capitalise the first letter of each token
	for i := range tokens {
		tokens[i] = strings.ToUpper(tokens[i][:1]) + tokens[i][1:]
	}
	return strings.Join(tokens, ` `)
}
