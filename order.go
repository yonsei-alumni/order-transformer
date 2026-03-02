package main

import (
	"fmt"
	"strconv"

	"github.com/xuri/excelize/v2"
)

// SourceOrder represents a single row from the source order Excel file.
type SourceOrder struct {
	OrderNumber    string
	Quantity       int
	ProductName    string
	OptionName     string
	RecipientName  string
	RecipientPhone string
	PostalCode     string
	Address        string
	DetailAddress  string
	DeliveryMemo   string
	CancelReason   string
}

// ShippingOrder represents a single row in the target shipping instruction Excel file.
type ShippingOrder struct {
	OrderNumber     string
	RecipientName   string
	FullAddress     string
	PostalCode      string
	Phone           string
	ItemCode        string
	Quantity        int
	DeliveryMessage string
	ProductName     string
}

// IsCancelled returns true if the order has a cancellation reason.
func (o SourceOrder) IsCancelled() bool {
	return o.CancelReason != ""
}

// CombinedProductName returns the product name combined with option name.
func (o SourceOrder) CombinedProductName() string {
	if o.OptionName == "" {
		return o.ProductName
	}
	return o.ProductName + " / " + o.OptionName
}

// CombinedAddress returns the full address combined from address and detail address.
func (o SourceOrder) CombinedAddress() string {
	if o.DetailAddress == "" {
		return o.Address
	}
	return o.Address + " " + o.DetailAddress
}

// Source Excel column indices (0-based).
const (
	srcColOrderNumber    = 1  // B: 주문번호
	srcColQuantity       = 16 // Q: 구매수량
	srcColProductName    = 17 // R: 상품명
	srcColOptionName     = 18 // S: 옵션명
	srcColRecipientName  = 24 // Y: 수령자명
	srcColRecipientPhone = 25 // Z: 수령자 전화번호
	srcColPostalCode     = 27 // AB: 배송지 우편번호
	srcColAddress        = 28 // AC: 주소
	srcColDetailAddress  = 29 // AD: 상세주소
	srcColDeliveryMemo   = 30 // AE: 배송메모
	srcColCancelReason   = 32 // AG: 취소사유
)

// ReadSourceOrders reads source order data from an Excel file.
func ReadSourceOrders(path string) ([]SourceOrder, error) {
	f, err := excelize.OpenFile(path)
	if err != nil {
		return nil, fmt.Errorf("엑셀 파일을 열 수 없습니다: %w", err)
	}
	defer f.Close()

	sheet := f.GetSheetName(0)
	rows, err := f.GetRows(sheet)
	if err != nil {
		return nil, fmt.Errorf("시트 데이터를 읽을 수 없습니다: %w", err)
	}

	if len(rows) < 2 {
		return nil, fmt.Errorf("데이터가 없습니다 (헤더만 존재)")
	}

	var orders []SourceOrder
	for i, row := range rows[1:] {
		if isEmptyRow(row) {
			continue
		}

		order := SourceOrder{
			OrderNumber:    getCell(row, srcColOrderNumber),
			ProductName:    getCell(row, srcColProductName),
			OptionName:     getCell(row, srcColOptionName),
			RecipientName:  getCell(row, srcColRecipientName),
			RecipientPhone: getCell(row, srcColRecipientPhone),
			PostalCode:     getCell(row, srcColPostalCode),
			Address:        getCell(row, srcColAddress),
			DetailAddress:  getCell(row, srcColDetailAddress),
			DeliveryMemo:   getCell(row, srcColDeliveryMemo),
			CancelReason:   getCell(row, srcColCancelReason),
		}

		qtyStr := getCell(row, srcColQuantity)
		if qtyStr != "" {
			qty, err := strconv.Atoi(qtyStr)
			if err != nil {
				return nil, fmt.Errorf("행 %d: 구매수량 변환 오류 ('%s'): %w", i+2, qtyStr, err)
			}
			order.Quantity = qty
		}

		orders = append(orders, order)
	}

	if len(orders) == 0 {
		return nil, fmt.Errorf("유효한 주문 데이터가 없습니다")
	}

	return orders, nil
}

// ReadShippingOrders reads shipping order data from a target Excel file.
func ReadShippingOrders(path string) ([]ShippingOrder, error) {
	f, err := excelize.OpenFile(path)
	if err != nil {
		return nil, fmt.Errorf("엑셀 파일을 열 수 없습니다: %w", err)
	}
	defer f.Close()

	sheet := f.GetSheetName(0)
	rows, err := f.GetRows(sheet)
	if err != nil {
		return nil, fmt.Errorf("시트 데이터를 읽을 수 없습니다: %w", err)
	}

	if len(rows) < 2 {
		return nil, fmt.Errorf("데이터가 없습니다 (헤더만 존재)")
	}

	var orders []ShippingOrder
	for i, row := range rows[1:] {
		if isEmptyRow(row) {
			continue
		}

		order := ShippingOrder{
			OrderNumber:     getCell(row, 0),
			RecipientName:   getCell(row, 1),
			FullAddress:     getCell(row, 2),
			PostalCode:      getCell(row, 3),
			Phone:           getCell(row, 4),
			ItemCode:        getCell(row, 5),
			DeliveryMessage: getCell(row, 7),
			ProductName:     getCell(row, 8),
		}

		qtyStr := getCell(row, 6)
		if qtyStr != "" {
			qty, err := strconv.Atoi(qtyStr)
			if err != nil {
				return nil, fmt.Errorf("행 %d: 수량 변환 오류 ('%s'): %w", i+2, qtyStr, err)
			}
			order.Quantity = qty
		}

		orders = append(orders, order)
	}

	return orders, nil
}

func getCell(row []string, index int) string {
	if index >= len(row) {
		return ""
	}
	return row[index]
}

func isEmptyRow(row []string) bool {
	for _, cell := range row {
		if cell != "" {
			return false
		}
	}
	return true
}
