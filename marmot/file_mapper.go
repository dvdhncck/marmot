package marmot

import (
	"log"
)

func MapLocation(album *Album) string {

	oldLocation := album.mediaFolder.folderPath

	newLocation := oldLocation + `_x`

	log.Printf("Xq %s, %s", oldLocation, newLocation)

	/*
	 things to fix....

	  leading /

	  upper case

	  subdirectory

	*/

	return newLocation
}
