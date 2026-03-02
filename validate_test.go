package main

import (
	"strings"
	"testing"
)

func TestValidate(t *testing.T) {
	sourceOrders, err := ReadSourceOrders("testdata/source-order-example.xlsx")
	if err != nil {
		t.Fatalf("ReadSourceOrders failed: %v", err)
	}

	shippingOrders := Transform(sourceOrders)
	result := Validate(sourceOrders, shippingOrders)

	if result.SourceTotalCount != 7 {
		t.Errorf("expected source total count 7, got %d", result.SourceTotalCount)
	}
	if result.SourceCancelledCount != 2 {
		t.Errorf("expected 2 cancelled orders, got %d", result.SourceCancelledCount)
	}
	if result.SourceValidCount != 5 {
		t.Errorf("expected 5 valid orders, got %d", result.SourceValidCount)
	}
	if result.SourceTotalQuantity != 5 {
		t.Errorf("expected source total quantity 5, got %d", result.SourceTotalQuantity)
	}
	if result.TargetOrderCount != 5 {
		t.Errorf("expected target order count 5, got %d", result.TargetOrderCount)
	}
	if result.TargetTotalQuantity != 5 {
		t.Errorf("expected target total quantity 5, got %d", result.TargetTotalQuantity)
	}

	if len(result.Errors) != 0 {
		t.Errorf("expected no errors, got: %v", result.Errors)
	}
}

func TestValidateProductBreakdown(t *testing.T) {
	sourceOrders, err := ReadSourceOrders("testdata/source-order-example.xlsx")
	if err != nil {
		t.Fatalf("ReadSourceOrders failed: %v", err)
	}

	shippingOrders := Transform(sourceOrders)
	result := Validate(sourceOrders, shippingOrders)

	expectedProducts := map[string]int{
		"연세25주년 재상봉 기념티셔츠 (Royal Blue) / 사이즈선택 : L": 2,
		"연세25주년 재상봉 기념티셔츠 (White) / 사이즈선택 : L":       2,
		"연세25주년 재상봉 기념티셔츠 (White) / 사이즈선택 : S":       1,
	}

	for product, expectedQty := range expectedProducts {
		if qty, ok := result.SourceProductBreakdown[product]; !ok || qty != expectedQty {
			t.Errorf("product '%s': expected %d, got %d", product, expectedQty, qty)
		}
	}
}

func TestFormatValidation(t *testing.T) {
	sourceOrders, err := ReadSourceOrders("testdata/source-order-example.xlsx")
	if err != nil {
		t.Fatalf("ReadSourceOrders failed: %v", err)
	}

	shippingOrders := Transform(sourceOrders)
	result := Validate(sourceOrders, shippingOrders)
	output := FormatValidation(result)

	if !strings.Contains(output, "전체 주문 건수: 7건") {
		t.Errorf("output should contain total order count")
	}
	if !strings.Contains(output, "취소 건수: 2건") {
		t.Errorf("output should contain cancelled count")
	}
	if !strings.Contains(output, "유효 주문 건수: 5건") {
		t.Errorf("output should contain valid order count")
	}
	if !strings.Contains(output, "모든 검증을 통과했습니다") {
		t.Errorf("output should contain success message")
	}
}

func TestValidateMismatch(t *testing.T) {
	sourceOrders := []SourceOrder{
		{OrderNumber: "001", Quantity: 2, ProductName: "Product A", RecipientName: "Kim"},
		{OrderNumber: "002", Quantity: 1, ProductName: "Product B", RecipientName: "Lee"},
	}

	targetOrders := []ShippingOrder{
		{OrderNumber: "001", Quantity: 1, ProductName: "Product A", RecipientName: "Kim"},
	}

	result := Validate(sourceOrders, targetOrders)

	if len(result.Errors) == 0 {
		t.Error("expected validation errors for mismatched data")
	}

	hasCountError := false
	hasQuantityError := false
	for _, e := range result.Errors {
		if strings.Contains(e, "주문 건수 불일치") {
			hasCountError = true
		}
		if strings.Contains(e, "상품 수량 불일치") {
			hasQuantityError = true
		}
	}

	if !hasCountError {
		t.Error("expected order count mismatch error")
	}
	if !hasQuantityError {
		t.Error("expected quantity mismatch error")
	}
}
