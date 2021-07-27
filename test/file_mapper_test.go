package main

import (
	"log"
	"testing"

	m "davidhancock.com/marmot/marmot"
)

func Test_shouldMapFileLocation(t *testing.T) {

	/*
	MediaFolder
	  archiveName string  <- deprecated, and of no value
	  mountPoint  string  <- meaningless
	  rootPath    string  00120, p2p#74, the_holding_pen, Jo 23, 1.4, v22
	  folderPath  string  <- unique, 1 per album (*)


	  (*) Except for various_artists__harthouse_8 and 
	      various_artists__easy_tempo_vol_1_a_cinematic_easy_listening_experience
		  both of which have 2 entries (2x AlbumID etc)
    */

	ins := []string { 
		`/this_has__leading_delimiter`, 
		`Thingy thing/The Backslashe\'s`,
		`one__two__three`,
		`one__two`,
		`one/two`, 
		`one\\two`, 
		`this_is_the_artist__this_is_the_album`,
		`Bing & Bong/Bangle Bingle #2`,
		`I "like" quotes/and so "do" I?`,
	}

	for _, in := range ins {

		album := m.NewAlbumFromDatabase(in)
		
		out, _ := m.MapLocation(album)
			
		log.Printf("%s\n", out)
	}
}