package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadSourceOrders(t *testing.T) {
	orders, err := ReadSourceOrders("testdata/source-order-example.xlsx")
	if err != nil {
		t.Fatalf("ReadSourceOrders failed: %v", err)
	}

	if len(orders) != 7 {
		t.Errorf("expected 7 orders, got %d", len(orders))
	}

	first := orders[0]
	if first.OrderNumber != "202504292700422" {
		t.Errorf("expected order number '202504292700422', got '%s'", first.OrderNumber)
	}
	if first.RecipientName != "김연세" {
		t.Errorf("expected recipient '김연세', got '%s'", first.RecipientName)
	}
	if first.Quantity != 1 {
		t.Errorf("expected quantity 1, got %d", first.Quantity)
	}
	if first.ProductName != "연세25주년 재상봉 기념티셔츠 (Royal Blue)" {
		t.Errorf("unexpected product name: '%s'", first.ProductName)
	}
	if first.OptionName != "사이즈선택 : L" {
		t.Errorf("unexpected option name: '%s'", first.OptionName)
	}
	if first.CancelReason != "" {
		t.Errorf("expected no cancel reason, got '%s'", first.CancelReason)
	}
}

func TestCancelledOrders(t *testing.T) {
	orders, err := ReadSourceOrders("testdata/source-order-example.xlsx")
	if err != nil {
		t.Fatalf("ReadSourceOrders failed: %v", err)
	}

	cancelledCount := 0
	for _, order := range orders {
		if order.IsCancelled() {
			cancelledCount++
		}
	}

	if cancelledCount != 2 {
		t.Errorf("expected 2 cancelled orders, got %d", cancelledCount)
	}

	// Row 3 (index 1): 이연세 - "구매 의사 취소"
	if orders[1].CancelReason != "구매 의사 취소" {
		t.Errorf("expected cancel reason '구매 의사 취소', got '%s'", orders[1].CancelReason)
	}

	// Row 8 (index 6): 최연세 - "입금전 취소"
	if orders[6].CancelReason != "입금전 취소" {
		t.Errorf("expected cancel reason '입금전 취소', got '%s'", orders[6].CancelReason)
	}
}

func TestTransform(t *testing.T) {
	orders, err := ReadSourceOrders("testdata/source-order-example.xlsx")
	if err != nil {
		t.Fatalf("ReadSourceOrders failed: %v", err)
	}

	shippingOrders := Transform(orders)

	if len(shippingOrders) != 5 {
		t.Errorf("expected 5 shipping orders (7 total - 2 cancelled), got %d", len(shippingOrders))
	}

	// Check first shipping order (김연세)
	first := shippingOrders[0]
	if first.OrderNumber != "202504292700422" {
		t.Errorf("expected order number '202504292700422', got '%s'", first.OrderNumber)
	}
	if first.RecipientName != "김연세" {
		t.Errorf("expected recipient '김연세', got '%s'", first.RecipientName)
	}
	if first.PostalCode != "06535" {
		t.Errorf("expected postal code '06535', got '%s'", first.PostalCode)
	}
	if first.Phone != "010-1234-5678" {
		t.Errorf("expected phone '010-1234-5678', got '%s'", first.Phone)
	}
	if first.FullAddress != "서울 서초구 반포대로 201 반포자이아파트 105동 1203호" {
		t.Errorf("unexpected address: '%s'", first.FullAddress)
	}
	if first.Quantity != 1 {
		t.Errorf("expected quantity 1, got %d", first.Quantity)
	}
	if first.ProductName != "연세25주년 재상봉 기념티셔츠 (Royal Blue) / 사이즈선택 : L" {
		t.Errorf("unexpected product name: '%s'", first.ProductName)
	}

	// Check delivery memo for 송연세 (index 4 in shipping orders)
	songOrder := shippingOrders[4]
	if songOrder.RecipientName != "송연세" {
		t.Errorf("expected recipient '송연세', got '%s'", songOrder.RecipientName)
	}
	if songOrder.DeliveryMessage != "배송 전에 미리 연락 바랍니다." {
		t.Errorf("unexpected delivery message: '%s'", songOrder.DeliveryMessage)
	}
}

func TestTransformAndWriteExcel(t *testing.T) {
	orders, err := ReadSourceOrders("testdata/source-order-example.xlsx")
	if err != nil {
		t.Fatalf("ReadSourceOrders failed: %v", err)
	}

	shippingOrders := Transform(orders)

	outputPath := filepath.Join(t.TempDir(), "output.xlsx")
	err = WriteShippingOrders(shippingOrders, outputPath)
	if err != nil {
		t.Fatalf("WriteShippingOrders failed: %v", err)
	}

	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatal("output file was not created")
	}

	// Read back the output and validate
	readBack, err := ReadShippingOrders(outputPath)
	if err != nil {
		t.Fatalf("ReadShippingOrders failed: %v", err)
	}

	if len(readBack) != len(shippingOrders) {
		t.Errorf("expected %d orders in output, got %d", len(shippingOrders), len(readBack))
	}

	if readBack[0].OrderNumber != shippingOrders[0].OrderNumber {
		t.Errorf("order number mismatch: %s != %s", readBack[0].OrderNumber, shippingOrders[0].OrderNumber)
	}
	if readBack[0].RecipientName != shippingOrders[0].RecipientName {
		t.Errorf("recipient mismatch: %s != %s", readBack[0].RecipientName, shippingOrders[0].RecipientName)
	}
}

func TestCombinedProductName(t *testing.T) {
	tests := []struct {
		product  string
		option   string
		expected string
	}{
		{"상품A", "옵션1", "상품A / 옵션1"},
		{"상품B", "", "상품B"},
	}

	for _, tc := range tests {
		order := SourceOrder{ProductName: tc.product, OptionName: tc.option}
		if got := order.CombinedProductName(); got != tc.expected {
			t.Errorf("CombinedProductName(%q, %q) = %q, want %q", tc.product, tc.option, got, tc.expected)
		}
	}
}

func TestCombinedAddress(t *testing.T) {
	tests := []struct {
		address string
		detail  string
		expected string
	}{
		{"서울 강남구", "101동 201호", "서울 강남구 101동 201호"},
		{"서울 강남구", "", "서울 강남구"},
	}

	for _, tc := range tests {
		order := SourceOrder{Address: tc.address, DetailAddress: tc.detail}
		if got := order.CombinedAddress(); got != tc.expected {
			t.Errorf("CombinedAddress(%q, %q) = %q, want %q", tc.address, tc.detail, got, tc.expected)
		}
	}
}
