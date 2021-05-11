package marmot

import (
	"log"
	"strings"
)

		 

func guessFromTags(album *Album) string {
	return transmute(album.artists[0].name) + `__` + transmute(album.name)
}

func transmute(input string) string {
	newLocation := input

	newLocation = strings.TrimLeft(newLocation, `/`)   // leading /
	
	newLocation = strings.ToLower(newLocation)   // case
		
	newLocation = strings.ReplaceAll(newLocation, `/`, `__`)  // unix path seperator

	newLocation = strings.ReplaceAll(newLocation, `\'`, ``)  // escaped apostrophe

	newLocation = strings.ReplaceAll(newLocation, ` & `, ` and `)
	newLocation = strings.ReplaceAll(newLocation, `_&_`, `_and_`)
	newLocation = strings.ReplaceAll(newLocation, `&`, ` and `)

	for strings.Contains(newLocation,`\\\\`) {
		newLocation = strings.ReplaceAll(newLocation, `\\\\`, `\\`)
	}

	newLocation = strings.ReplaceAll(newLocation, `\\`, `__`)  // windows path seperator
	newLocation = strings.ReplaceAll(newLocation, `\`, `__`)  // windows path seperator
	
	newLocation = strings.ReplaceAll(newLocation, `'`, ``)   // apostrophe

	newLocation = strings.ReplaceAll(newLocation, `- `, ``) // dash-space
	newLocation = strings.ReplaceAll(newLocation, `-`, ``)  // dash
	

	// collapse double spaces recursively
	for strings.Contains(newLocation,`  `) {
		newLocation = strings.ReplaceAll(newLocation, `  `, ` `)
	}

	newLocation = strings.ReplaceAll(newLocation, ` _`, `_`) // space-underscore

	newLocation = strings.ReplaceAll(newLocation, ` `, `_`)  // space

	return newLocation
} 
	
func MapLocation(album *Album) (string, int) {

	oldLocation := album.mediaFolder.folderPath

	newLocation := transmute(oldLocation)

	mapState := NO_CHANGE

	if newLocation != oldLocation {
		mapState = GOOD_MAP
	}

	if strings.Count(newLocation,`__`) > 1 {
		log.Printf("PROBLEM: %s -> %s\n", oldLocation, newLocation)
		mapState = PROBLEM_MAP
	}

	if strings.Count(newLocation,`__`) == 0 {
		tagGuess := guessFromTags(album)
		log.Printf("FAIL: %s -> %s -> %s\n", oldLocation, newLocation, tagGuess)
		newLocation = tagGuess
		mapState = MAP_FAIL
	}

	return newLocation, mapState
}
