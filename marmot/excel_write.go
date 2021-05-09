package marmot

import (
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize/v2"
)

func ExcelWrite(fileName string, collection Collection) {

	f := excelize.NewFile()

	sheet := `Flarp`
	index := f.NewSheet(sheet)


	styleA, err := f.NewStyle(`{"fill":{"type":"pattern","color":["#dd1111"],"pattern":1}}`)
	if err != nil {
		fmt.Println(err)
	}
	styleB, err := f.NewStyle(`{"fill":{"type":"pattern","color":["#11dd11"],"pattern":1}}`)
	if err != nil {
		fmt.Println(err)
	}
	styleC, err := f.NewStyle(`{"fill":{"type":"pattern","color":["#1111dd"],"pattern":1}}`)
	if err != nil {
		fmt.Println(err)
	}
	
	row := 1
	for albumId, album := range collection.lookup {
		f.SetRowHeight(sheet, row, 20)
		f.SetCellValue(sheet, fmt.Sprintf(`A%d`, row), albumId)
		f.SetCellValue(sheet, fmt.Sprintf(`B%d`, row), album.name)
		f.SetCellValue(sheet, fmt.Sprintf(`C%d`, row), album.artists[0].name)
		f.SetCellValue(sheet, fmt.Sprintf(`D%d`, row), album.mediaFolder.rootPath)
		f.SetCellValue(sheet, fmt.Sprintf(`E%d`, row), album.mediaFolder.folderPath)

		style := styleA
		if row % 3 == 0 { style = styleB }
		if row % 7 == 0 { style = styleC }

		f.SetCellStyle(sheet, fmt.Sprintf(`A%d`, row), fmt.Sprintf(`E%d`, row), style)

		row++
	}

	f.SetColWidth(sheet, "A", "A", 8)
	f.SetColWidth(sheet, "B", "B", 40)
	f.SetColWidth(sheet, "C", "C", 30)
	f.SetColWidth(sheet, "D", "D", 10)
	f.SetColWidth(sheet, "E", "E", 120)

	f.SetActiveSheet(index)

	if err := f.SaveAs(fileName); err != nil {
		fmt.Println(err)
	}
}
