package marmot

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

/*
	[ 'a': []]
*/


func NewGenreButler() *GenreButler {
	toings := make(map[string][]*Metadata, 2048)
	froings := make(map[int64]*Metadata, 2048)
	return &GenreButler{&toings, &froings, &GenreForestNode{``,make([]*GenreForestNode, 0)}}
}

type GenreButler struct {
	genrePathToMetadataList *map[string][]*Metadata
	albumIdToMetadata       *map[int64]*Metadata
	genreForest             *GenreForestNode
}

func (genreButler *GenreButler) transformMetaFile(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}
	if !info.IsDir() && filepath.Base(path) == `meta.json` {
		//log.Println(path)

		data, err := ioutil.ReadFile(path)

		if err != nil {
			return err
		}

		metadata := Metadata{}
		err = json.Unmarshal([]byte(data), &metadata)
		if err != nil {
			return err
		}

		// do whatever here

		if err != nil {
			return err
		}

		jsonRep, err := json.Marshal(&metadata)
		if err != nil {
			return err
		}

		err = ioutil.WriteFile(path, jsonRep, 0644)

		log.Println(fmt.Printf("Meta2: %v", jsonRep))

		if err != nil {
			return err
		}
	}
	return err
}

// a.b.c : newNode = child['a'] on genreForestNode, create if required
//         recursive "update(newNode, 'b.c')"

func (genreButler *GenreButler) updateGenreForest(genreForestNode *GenreForestNode, genrePath string) {
	// termination condition
	if genrePath == `` {
		return
	}
	
	var head string
	var tail string

	index := strings.Index(genrePath, ".")

	if index < 0 {
		head = genrePath
		tail = ``

	} else {
		head = genrePath[0:index]
		tail = genrePath[index+1:]		
	}

	// already in children of node?
	found := false
	for _, child := range genreForestNode.Children {
		if child.Name == head {
			found = true
			genreButler.updateGenreForest(child, tail)
		}
	}
	// not in children, add it
	if !found {
		child := GenreForestNode{head, make([]*GenreForestNode, 0)}
		genreForestNode.Children = append(genreForestNode.Children, &child)
		genreButler.updateGenreForest(&child, tail)
	}
}

func (genreButler *GenreButler) assimilate(libraryPath string, metadata *Metadata) {

	genrePathList := strings.Split(metadata.GenrePaths, ",")

	for _, genrePath := range genrePathList {

		list, exists := (*genreButler.genrePathToMetadataList)[genrePath]
		if !exists {
			list = []*Metadata{}
		}

		genreButler.updateGenreForest(genreButler.genreForest, genrePath)

		(*genreButler.genrePathToMetadataList)[genrePath] = append(list, metadata)

		//log.Println(fmt.Sprintf(">> %v %v", libraryPath, genrePath))
	}

	(*genreButler.albumIdToMetadata)[metadata.ID] = metadata
}

func (genreButler *GenreButler) gatherMetaData(metaFilePath string, info os.FileInfo, err error) error {
	if err == nil {
		if !info.IsDir() && filepath.Base(metaFilePath) == `meta.json` {
			data, err := ioutil.ReadFile(metaFilePath)
			if err == nil {
				metadata := Metadata{}
				err = json.Unmarshal([]byte(data), &metadata)
				if err == nil {
					// get root folder for file
					rootAlbumPath := filepath.Dir(metaFilePath)

					// if metadata.Location != rootAlbumPath {
					// 	log.Println(fmt.Sprintf("WARN: incorrect metadata.Location - is %v, should be %v", metadata.Location, rootAlbumPath))
					// }

					metadata.Location = rootAlbumPath // auto-fix location (not permnanently though)
					metadata.UrlBase = strings.Replace(rootAlbumPath, `/library`, ``, 1)

					genreButler.assimilate(rootAlbumPath, &metadata)
				}
			}
		}
	}
	return err
}

func (genreButler *GenreButler) GetRootGenres() []string {
	var m map[string]bool = make(map[string]bool, 128)
	
	for candidate := range *genreButler.genrePathToMetadataList {
		nodes := strings.Split(candidate, ".")
		m[nodes[0]] = true
	}
	keys := []string{}
	for k := range m {
		keys = append(keys, k)
	}

	return keys
}

func (genreButler *GenreButler) GetSubGenres(genrePath string) []string {
	var m map[string]bool = make(map[string]bool, 128)
	
	for candidate := range *genreButler.genrePathToMetadataList {
		if candidate != genrePath {
			if strings.HasPrefix(strings.ToLower(candidate), strings.ToLower(genrePath)) {
				m[candidate] = true
			}
		}
	}
	keys := []string{}
	for k := range m {
		keys = append(keys, k)
	}

	return keys
}

func (genreButler *GenreButler) GetAlbumsForGenrePath(targetGenrePath string) ([]*Metadata, error) {
	var result []*Metadata

	if strings.ToLower(targetGenrePath) == `new` {
		for _, metadata := range *genreButler.albumIdToMetadata {
			result = append(result, metadata)
		}
		// sort by ID, highest first
		sort.Slice(result, func(i, j int) bool { return result[i].ID > result[j].ID })
		result = result[:50]

	} else if strings.ToLower(targetGenrePath) == `random` {
		result = make([]*Metadata, 0, len(*genreButler.albumIdToMetadata))
		for _, metadata := range *genreButler.albumIdToMetadata {
			if len(result) < 50 {
				result = append(result, metadata)
			}
		}
	} else {
		for genrePath, list := range *genreButler.genrePathToMetadataList {
			if strings.HasPrefix(strings.ToLower(genrePath), strings.ToLower(targetGenrePath)) {
				for _, metadata := range list {
					result = append(result, metadata)
				}
			}
		}
	}

	//random shuffle
	rand.Seed(time.Now().UnixNano())
	for i := range result {
		j := rand.Intn(i + 1)
		result[i], result[j] = result[j], result[i]
	}

	return result, nil
}

func (genreButler *GenreButler) GetMetadataByPath(path string) (*Metadata, error) {
	for _, metadata := range *genreButler.albumIdToMetadata {
		if path == metadata.Location {
			return metadata, nil
		}
	}
	return nil, errors.New(fmt.Sprintf("Unknown path %v", path))
}

func (genreButler *GenreButler) GetMetadataById(albumId int64) (*Metadata, error) {
	metadata, exists := (*genreButler.albumIdToMetadata)[albumId]
	if exists {
		return metadata, nil
	} else {
		return nil, errors.New(fmt.Sprintf("No metadata for album %v", albumId))
	}
}

func (genreButler *GenreButler) GetAlbumsForText(text string) ([]*Metadata, error) {
	var result []*Metadata

	textL := strings.ToLower(text)
	for _, metadata := range *genreButler.albumIdToMetadata {
		if strings.Contains(strings.ToLower(metadata.Title), textL) || inStringSlice(metadata.Artists, textL) {
			result = append(result, metadata)
		}
	}

	return result, nil
}

func (genreButler *GenreButler) ScanLibrary() {
	start := time.Now()

	err := filepath.Walk(settings.rootPath, genreButler.gatherMetaData)

	if err != nil {
		log.Fatal(err)
	}

	log.Println(fmt.Sprintf("Visited %v albums in %v", len(*genreButler.albumIdToMetadata), time.Since(start)))
}


func (genreButler *GenreButler) ListAlbumsByGenre(targetGenrePath string) {
	for genrePath, list := range *genreButler.genrePathToMetadataList {
		if strings.HasPrefix(genrePath, targetGenrePath) {
			fmt.Printf("%v\n", genrePath)
			for _, metadata := range list {
				fmt.Printf("\t%v\n", metadata.Location)
			}
		}
	}
}

func (genreButler *GenreButler) ListGenreForest() {
	listGenreForest(genreButler.genreForest, ``)
}

func listGenreForest(genreForestNode *GenreForestNode, indent string) {
	if len(genreForestNode.Name) > 0 { 
		fmt.Printf("%s%s\n", indent, genreForestNode.Name)
	}
	for _, child := range genreForestNode.Children {
		listGenreForest(child, indent + "  ")
	}
}

func (genreButler *GenreButler) ListAllAlbums() {

	var result []*Metadata
	for _, metadata := range *genreButler.albumIdToMetadata {
			result = append(result, metadata)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Location < result[j].Location })
		
	for _, metadata := range result {
		genrePathList := strings.Split(metadata.GenrePaths, ",")
		for _, genrePath := range genrePathList {
			fmt.Printf("%v\t%v\n", metadata.Location, genrePath)
		}
	}
}

func inStringSlice(haystack []string, needle string) bool {
	for _, hay := range haystack {
		if strings.Contains(strings.ToLower(hay), needle) {
			return true
		}
	}
	return false
}
