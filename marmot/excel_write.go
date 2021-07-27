package marmot

import (
	"fmt"
	"log"
	"github.com/360EntSecGroup-Skylar/excelize/v2"
)

func ExcelWrite(fileName string, collection Collection) {

	f := excelize.NewFile()
	index := f.GetActiveSheetIndex();
	sheet := f.GetSheetName(index)

	// sheet := `Flarp`
	// index := f.NewSheet(sheet)

	goodMapStyle, err := f.NewStyle(`{"fill":{"type":"pattern","color":["#e0c880"],"pattern":1}}`)
	if err != nil {
		log.Fatal(err)
	}
	problemMapStyle, err := f.NewStyle(`{"fill":{"type":"pattern","color":["#abe080"],"pattern":1}}`)
	if err != nil {
		log.Fatal(err)
	}
	guessedMapStyle, err := f.NewStyle(`{"fill":{"type":"pattern","color":["#ce7feb"],"pattern":1}}`)
	if err != nil {
		log.Fatal(err)
	}
	

	row := 1
	for albumId, album := range collection.inDatabase {
		f.SetRowHeight(sheet, row, 20)
		f.SetCellValue(sheet, fmt.Sprintf(`A%d`, row), albumId)
		f.SetCellValue(sheet, fmt.Sprintf(`B%d`, row), album.name)
		f.SetCellValue(sheet, fmt.Sprintf(`C%d`, row), album.artists[0].name)
		f.SetCellValue(sheet, fmt.Sprintf(`D%d`, row), album.mediaFolder.rootPath)
		f.SetCellValue(sheet, fmt.Sprintf(`E%d`, row), album.mediaFolder.folderPath)
		f.SetCellValue(sheet, fmt.Sprintf(`F%d`, row), album.mapState)
		f.SetCellValue(sheet, fmt.Sprintf(`G%d`, row), album.location)

		if album.mapState == GOOD_MAP {
			f.SetCellStyle(sheet, fmt.Sprintf(`A%d`, row), fmt.Sprintf(`G%d`, row), goodMapStyle)
		}
		if album.mapState == PROBLEM_MAP {
			f.SetCellStyle(sheet, fmt.Sprintf(`A%d`, row), fmt.Sprintf(`G%d`, row), problemMapStyle)
		}
		if album.mapState == MAP_FAIL {
			f.SetCellStyle(sheet, fmt.Sprintf(`A%d`, row), fmt.Sprintf(`G%d`, row), guessedMapStyle)
		}
		row++
	}

	f.SetColWidth(sheet, "A", "A", 8)
	f.SetColWidth(sheet, "B", "B", 40)
	f.SetColWidth(sheet, "C", "C", 30)
	f.SetColWidth(sheet, "D", "D", 10)
	f.SetColWidth(sheet, "E", "E", 100)
	f.SetColWidth(sheet, "F", "F", 4)
	f.SetColWidth(sheet, "G", "G", 100)

	f.SetActiveSheet(index)

	if err := f.SaveAs(fileName); err != nil {
		log.Fatal(err)
	}
}
