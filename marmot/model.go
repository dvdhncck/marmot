package marmot

import (
	"fmt"
	"strings"
)

type MediaFolder struct {
	archiveName string
	mountPoint  string
	rootPath    string
	folderPath  string
}

func (mf *MediaFolder) ToJson() string {
	return fmt.Sprintf("[\"%s/%s/%s\"]",
		mf.mountPoint, mf.rootPath, mf.folderPath)
}

type Artist struct {
	id     string
	name   string
	sortAs string
}

type Album struct {
	id          string
	name        string
	mapState    int
	mediaFolder *MediaFolder
	location    string          // this is where we will migrate it to
	sortAs      string
	artists     []*Artist
	genres      []string
}

const NO_CHANGE = 0
const GOOD_MAP = 1
const PROBLEM_MAP = 2
const MAP_FAIL = 3

func NewAlbum(folderPath string) *Album {
	a := Album{}
	a.mediaFolder = &MediaFolder{}
	a.mediaFolder.folderPath = folderPath
	return &a
}

func (album *Album) GetOldLocation() *MediaFolder { return album.mediaFolder }

func (album *Album) ToJson() string {
	artists := []string{}
	for _, artist := range album.artists {
		artists = append(artists, fmt.Sprintf("\"%s\"", artist.name))
	}
	genres := []string{}
	for _, genre := range album.genres {
		genres = append(genres, fmt.Sprintf("\"%s\"", genre))
	}
	return fmt.Sprintf("{ id: \"%s\",\n  name: \"%s\",\n  oldLocation: %s,\n  newLocation: %s,\n  sortAs: \"%s\",\n  genres: [%s]\n  artists: [%s]\n}",
		album.id, album.name, album.mediaFolder.ToJson(), album.location, album.sortAs, strings.Join(genres, `,`), strings.Join(artists, `,`))
}