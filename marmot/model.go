package marmot

import (
	"fmt"
	"sort"
)

// DEPRECATED
type Genre struct {
	id     string
	name   string
}

// DEPRECATED
type Artist struct {
	id     string
	name   string
	sortAs string
}

// DEPRECATED
type Album struct {
	id          string
	name        string
	location    string 
	sortAs      string
	artists     []*Artist
	genres      []*Genre
}

// DEPRECATED
type MinimalAlbum struct {
	Id string         `json:"id"`        // primary key for subsequent searches
	Location string   `json:"location"`  // sufficient to construct the cover image url
}

// DEPRECATED
type AlbumMetadata struct {
	Id string         `json:"id"`        // primary key for subsequent searches
	Location string   `json:"location"`  // sufficient to construct the cover image url
	Title string      `json:"title"`
	Artists string    `json:"artists"`
}

type Track struct {
	Number  int       `json:"number"`
	Title   string    `json:"title"`
	Artist  string    `json:"artist"`
	Url     string    `json:"url"`
}

type Metadata struct {
	ID 		int64       `json:"id"`
	Title   string      `json:"title"`
	Location string     `json:"location"`
	UrlBase string      `json:"urlBase"`
	Genres  []string    `json:"genres"`
	GenrePaths  string  `json:"genrePaths"` // can be empty
	Artists []string    `json:"artists"`
}

type Playlist struct {
	Metadata  *Metadata  `json:"metadata"`
	Tracks    []*Track   `json:"tracks"`
}

type GenreForestNode struct {
	Name string                 `json:"name"`
	Children []*GenreForestNode `json:"children"`
}

func DEPRECATED_NewPlaylist(metadata *Metadata, tracks []*Track) *Playlist {
  p := Playlist{}
  p.Metadata = metadata
  p.Tracks = tracks
  sort.Slice(p.Tracks, func (i,j int) bool { return tracks[i].Url < tracks[j].Url })
  return &p
}

func DEPRECATED_NewAlbumFromFilesystem(id int64, location string, title string, artists []string, genres []string) *Album {
	a := Album{}
	a.id = fmt.Sprintf("%d",id)
	a.name = title
	a.artists = []*Artist{}
	for _, artist := range artists {
		a.artists = append(a.artists, &Artist{``, artist, ``})
	}
	for _, genre := range genres {
		a.genres = append(a.genres, &Genre{``, genre})
	}
	a.location = location
	return &a
}