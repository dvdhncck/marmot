package marmot

import (
	"fmt"
	"sort"
)

type Genre struct {
	id     string
	name   string
}

type Artist struct {
	id     string
	name   string
	sortAs string
}

type Album struct {
	id          string
	name        string
	location    string 
	sortAs      string
	artists     []*Artist
	genres      []*Genre
}

type MinimalAlbum struct {
	Id string         `json:"id"`        // primary key for subsequent searches
	Location string   `json:"location"`  // sufficient to construct the cover image url
}

type Track struct {
	Name    string    `json:"name"`
	Artist  string    `json:"artist"`
	File    string    `json:"file"`
	Url     string    `json:"url"`
}

type Playlist struct {
	AlbumID string    `json:"albumId"`
	Title   string    `json:"title"`
	Tracks  []*Track  `json:"tracks"`
}

type Metadata struct {
	ID 		int64     `json:"id"`
	Title   string    `json:"title"`
	Genres  []string  `json:"genres"`
	Artists []string  `json:"artists"`
}

func NewPlaylist(albumId string, title string, tracks []*Track) *Playlist {
  p := Playlist{}
  p.AlbumID = albumId
  p.Title = title
  p.Tracks = tracks
  sort.Slice(p.Tracks, func (i,j int) bool { return tracks[i].File < tracks[j].File })
  return &p
}

func NewAlbumFromFilesystem(id int64, location string, title string, artists []string, genres []string) *Album {
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