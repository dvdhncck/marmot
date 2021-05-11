package marmot

import (
	"log"
	"github.com/360EntSecGroup-Skylar/excelize/v2"
)

func ExcelRead(fileName string, collection Collection) {

	f, err := excelize.OpenFile(fileName)

	if err != nil {
		log.Fatal(err)
	}

	index := f.GetActiveSheetIndex()
	sheet := f.GetSheetName(index)
	rows, _ := f.GetRows(sheet)

	var albumId, newLocation string

	count:= 0
	for _, row := range rows {
		albumId = row[0]
		newLocation = row[6]
		if albumId != `` && newLocation != `` {
			err := collection.UpdateNewLocation(albumId, newLocation)
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("Updated: %s -> %s\n", albumId, newLocation)
			count += 1
		}
	}
	log.Printf("Updated %d items in the collection\n", count)

}
