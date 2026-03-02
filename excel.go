package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xuri/excelize/v2"
)

const targetSheetName = "이지어드민 양식"

// WriteShippingOrders writes shipping orders to a target Excel file.
func WriteShippingOrders(orders []ShippingOrder, outputPath string) error {
	f := excelize.NewFile()
	defer f.Close()

	f.SetSheetName("Sheet1", targetSheetName)

	headers := []string{
		"주문번호", "수령인", "수령인 주소(전체)", "수령인 우편번호",
		"수령인 휴대전화", "자체품목코드", "수량", "배송메시지", "주문상품명(옵션포함)",
	}

	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(targetSheetName, cell, h)
	}

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})
	f.SetCellStyle(targetSheetName, "A1", "I1", headerStyle)

	for i, order := range orders {
		row := i + 2
		f.SetCellValue(targetSheetName, fmt.Sprintf("A%d", row), order.OrderNumber)
		f.SetCellValue(targetSheetName, fmt.Sprintf("B%d", row), order.RecipientName)
		f.SetCellValue(targetSheetName, fmt.Sprintf("C%d", row), order.FullAddress)
		f.SetCellValue(targetSheetName, fmt.Sprintf("D%d", row), order.PostalCode)
		f.SetCellValue(targetSheetName, fmt.Sprintf("E%d", row), order.Phone)
		f.SetCellValue(targetSheetName, fmt.Sprintf("F%d", row), order.ItemCode)
		f.SetCellValue(targetSheetName, fmt.Sprintf("G%d", row), order.Quantity)
		f.SetCellValue(targetSheetName, fmt.Sprintf("H%d", row), order.DeliveryMessage)
		f.SetCellValue(targetSheetName, fmt.Sprintf("I%d", row), order.ProductName)
	}

	f.SetColWidth(targetSheetName, "A", "A", 20)
	f.SetColWidth(targetSheetName, "B", "B", 12)
	f.SetColWidth(targetSheetName, "C", "C", 40)
	f.SetColWidth(targetSheetName, "D", "D", 15)
	f.SetColWidth(targetSheetName, "E", "E", 18)
	f.SetColWidth(targetSheetName, "F", "F", 15)
	f.SetColWidth(targetSheetName, "G", "G", 8)
	f.SetColWidth(targetSheetName, "H", "H", 30)
	f.SetColWidth(targetSheetName, "I", "I", 45)

	return f.SaveAs(outputPath)
}

// uniqueFilename returns a unique filename by appending (1), (2), etc.
// if the file already exists, like browser download behavior.
func uniqueFilename(basePath string) string {
	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		return basePath
	}

	ext := filepath.Ext(basePath)
	name := strings.TrimSuffix(basePath, ext)

	for i := 1; ; i++ {
		candidate := fmt.Sprintf("%s(%d)%s", name, i, ext)
		if _, err := os.Stat(candidate); os.IsNotExist(err) {
			return candidate
		}
	}
}
