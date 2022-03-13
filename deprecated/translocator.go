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

func (collection *Collection) Translocate(db *sql.DB) {

	failed := []string{}
	for _, album := range collection.inDatabase {

		source := fmt.Sprintf("%s/%s", album.mediaFolder.rootPath, album.mediaFolder.folderPath)
		source = strings.ReplaceAll(source, `\`, `/`)

		destination := album.location

		log.Println("Translocating: ", source, " to ", destination)

		if strings.Count(destination, `___`) != 0 {
			failed = append(failed, "[DODGY_DELIM] "+destination+" id:"+album.id)
		} else {
			if strings.Count(destination, `\`) != 0 || strings.Count(destination, `/`) != 0 {
				failed = append(failed, "[DODGY_DEST] "+destination)
			} else {
				if strings.HasPrefix(source, "Jo 14") {
					log.Println("Skipping Jo 14")
					failed = append(failed, "[SKIPPED] "+source)
				} else {
					if _, err := os.Stat(source); err == nil {

						err = validateFolder(source)

						if err == nil {
							flyDest := filepath.Join(`fly`, destination)
							err = os.MkdirAll(flyDest, 0777)
							if err != nil {
								failed = append(failed, "[MKDIR_FAILED] "+flyDest)
							} else {

								err = CopyCoverArt(album, flyDest)
								if err != nil {
									failed = append(failed, "[COVER_ART] "+err.Error())
								}

								err = WriteMetadata(db, album, flyDest)
								if err != nil {
									failed = append(failed, "[WRITE_META_FAILED] "+flyDest)
								} else {
									fails := NewFails(flyDest)
									validateMetadata(fails, flyDest)
									if fails.IsGood() {
										MoveTracks(source, flyDest)
									} else {
										failed = append(failed, "[META_VALIDATION_FAILED] "+flyDest)
										fails.Write()
									}
								}

								// metadata and cover ok, move the source files
								
							}

							// if !settings.dryRun {
							// 	err := os.Rename(source, destination)
							// 	if err != nil {
							// 		log.Print(err)
							// 		failed = append(failed, source)
							// 	}
							// }

						} else {
							failed = append(failed, err.Error()+` `+source)
						}
					} else {
						log.Println("Source does not exist: ", source)
						failed = append(failed, source)
					}
				}
			}
		}
	}

	for _, fail := range failed {
		log.Println("Failed to move: ", fail)
	}

}
