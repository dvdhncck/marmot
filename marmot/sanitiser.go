package marmot

import (
	"database/sql"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

/*
func validateFolder(path string) error {
	pathFileInfo, err := ioutil.ReadDir(path)
	// should contain at least the metadata file, and 0 directories
	if err == nil {
		if len(pathFileInfo) == 0 {
			return errors.New("[NO FILES] " + path)
		} else {
			count := countSubDirsIn(pathFileInfo)
			if count > 0 {
				// all directories
				return errors.New("[ILLEGAL SUBDIRS] " + path)
			} else {
				return nil
			}
		}

	} else {
		return err
	}
}

func ensafen(in string) string {
	return `"` + strings.ReplaceAll(in, `"`, `\"`) + `"`
}
func ensafenGenres(in []*Genre) string {
	result := ""
	delim := "["
	for _, x := range in {
		result = result + delim + ensafen(x.name)
		delim = ","
	}
	return result + "]"
}
func ensafenArtists(in []*Artist) string {
	result := ""
	delim := "["
	for _, x := range in {
		result = result + delim + ensafen(x.name)
		delim = ","
	}
	return result + "]"
}

func WriteMetadata(db *sql.DB, album *Album, path string) error {
	json := fmt.Sprintf(
		`{
   "id" : %s,
   "title": %s,
   "genres": %s,
   "artists": %s
}
`, album.id, ensafen(album.name), ensafenGenres(album.genres), ensafenArtists(album.artists))

	return ioutil.WriteFile(filepath.Join(path, "meta.json"), []byte(json), 0644)

}

func CopyFile(sourceFile, destinationFile string) error {
	input, err := ioutil.ReadFile(sourceFile)
	if err != nil {
		fmt.Println(err)
		return err
	}

	err = ioutil.WriteFile(destinationFile, input, 0644)
	if err != nil {
		fmt.Println("Error creating", destinationFile)
		return err
	}
	return nil
}

func CopyCoverArt(album *Album, path string) error {
	sourcePath := "/home/dave/projects/html/marmot/art/" + album.id + "_album_cover.jpg"
	destinationPath := filepath.Join(path, "cover.jpg")
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		return errors.New("Failed to find " + sourcePath)
	} else {
		if strings.Count(destinationPath, ".jpg") != 1 {
			return errors.New("Dodge filename " + destinationPath)
		} else {			
			return CopyFile(sourcePath, destinationPath)
		}
	}
}

func MoveTracks(sourcePath, destinationPath string) error {
	contentsFileInfos, err := ioutil.ReadDir(sourcePath)
	if err == nil {
		for _, fileInfo := range contentsFileInfos {	
			sourceFile := filepath.Join(sourcePath, fileInfo.Name())
			destFile := filepath.Join(destinationPath, fileInfo.Name())
			err := os.Rename(sourceFile, destFile)
			if err != nil {
				return err
			}
			fmt.Printf("%s --> %s\n", sourceFile, destFile)
		}
		return nil
	} else {
		return err
	}
}
*/

func (collection *Collection) Sanitise(db *sql.DB) {

	failed := []string{}
	for _, album := range collection.inDatabase {

		
	}

	// for _, fail := range failed {
	// 	log.Println("Failed to move: ", fail)
	// }

}
