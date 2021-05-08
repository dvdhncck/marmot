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

func (mf *MediaFolder) toJson() string {
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
	mediaFolder *MediaFolder
	newLocation string          // this is where we will migrate it to
	sortAs      string
	artists     []*Artist
	genres      []string
}

func (album *Album) toJson() string {
	artists := []string{}
	for _, artist := range album.artists {
		artists = append(artists, fmt.Sprintf("\"%s\"", artist.name))
	}
	genres := []string{}
	for _, genre := range album.genres {
		genres = append(genres, fmt.Sprintf("\"%s\"", genre))
	}
	return fmt.Sprintf("{ id: \"%s\",\n  name: \"%s\",\n  oldLocation: %s,\n  oldLocation: %s,\n  sortAs: \"%s\",\n  genres: [%s]\n  artists: [%s]\n}",
		album.id, album.name, album.mediaFolder.toJson(), album.newLocation, album.sortAs, strings.Join(genres, `,`), strings.Join(artists, `,`))
}