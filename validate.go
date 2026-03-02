package main

import (
	"fmt"
	"sort"
	"strings"
)

// ValidationResult holds the result of validating source orders against shipping orders.
type ValidationResult struct {
	SourceTotalCount       int
	SourceCancelledCount   int
	SourceValidCount       int
	SourceTotalQuantity    int
	SourceProductBreakdown map[string]int

	TargetOrderCount       int
	TargetTotalQuantity    int
	TargetProductBreakdown map[string]int

	Errors   []string
	Warnings []string
}

// Validate compares source orders and shipping orders, returning a validation result.
func Validate(sourceOrders []SourceOrder, shippingOrders []ShippingOrder) ValidationResult {
	result := ValidationResult{
		SourceProductBreakdown: make(map[string]int),
		TargetProductBreakdown: make(map[string]int),
	}

	result.SourceTotalCount = len(sourceOrders)
	for _, order := range sourceOrders {
		if order.IsCancelled() {
			result.SourceCancelledCount++
			continue
		}
		result.SourceValidCount++
		result.SourceTotalQuantity += order.Quantity
		result.SourceProductBreakdown[order.CombinedProductName()] += order.Quantity
	}

	result.TargetOrderCount = len(shippingOrders)
	for _, order := range shippingOrders {
		result.TargetTotalQuantity += order.Quantity
		result.TargetProductBreakdown[order.ProductName] += order.Quantity
	}

	// Validate order counts
	if result.SourceValidCount != result.TargetOrderCount {
		result.Errors = append(result.Errors,
			fmt.Sprintf("주문 건수 불일치: 구매자 주문 %d건 ≠ 출고 지시 %d건",
				result.SourceValidCount, result.TargetOrderCount))
	}

	// Validate total quantities
	if result.SourceTotalQuantity != result.TargetTotalQuantity {
		result.Errors = append(result.Errors,
			fmt.Sprintf("상품 수량 불일치: 구매자 주문 %d개 ≠ 출고 지시 %d개",
				result.SourceTotalQuantity, result.TargetTotalQuantity))
	}

	// Validate product breakdowns
	allProducts := make(map[string]bool)
	for k := range result.SourceProductBreakdown {
		allProducts[k] = true
	}
	for k := range result.TargetProductBreakdown {
		allProducts[k] = true
	}

	for product := range allProducts {
		srcQty := result.SourceProductBreakdown[product]
		tgtQty := result.TargetProductBreakdown[product]
		if srcQty != tgtQty {
			result.Errors = append(result.Errors,
				fmt.Sprintf("상품별 수량 불일치 [%s]: 구매자 주문 %d개 ≠ 출고 지시 %d개",
					product, srcQty, tgtQty))
		}
	}

	// Additional validation: check for missing data in shipping orders
	for _, order := range shippingOrders {
		if order.RecipientName == "" {
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("주문번호 %s: 수령인 이름이 비어있습니다", order.OrderNumber))
		}
		if order.Phone == "" {
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("주문번호 %s: 수령인 전화번호가 비어있습니다", order.OrderNumber))
		}
		if order.FullAddress == "" {
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("주문번호 %s: 수령인 주소가 비어있습니다", order.OrderNumber))
		}
		if order.PostalCode == "" {
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("주문번호 %s: 우편번호가 비어있습니다", order.OrderNumber))
		}
	}

	return result
}

// FormatValidation formats the validation result as a human-readable string.
func FormatValidation(result ValidationResult) string {
	var sb strings.Builder

	sb.WriteString("=== 검증 결과 ===\n\n")

	sb.WriteString("[구매자 주문 요약]\n")
	sb.WriteString(fmt.Sprintf("  전체 주문 건수: %d건\n", result.SourceTotalCount))
	sb.WriteString(fmt.Sprintf("  취소 건수: %d건\n", result.SourceCancelledCount))
	sb.WriteString(fmt.Sprintf("  유효 주문 건수: %d건\n", result.SourceValidCount))
	sb.WriteString(fmt.Sprintf("  전체 상품 수량: %d개\n", result.SourceTotalQuantity))
	sb.WriteString("\n")

	sb.WriteString("[출고 지시 요약]\n")
	sb.WriteString(fmt.Sprintf("  주문 건수: %d건\n", result.TargetOrderCount))
	sb.WriteString(fmt.Sprintf("  전체 상품 수량: %d개\n", result.TargetTotalQuantity))
	sb.WriteString("\n")

	sb.WriteString("[상품 종류별 수량 - 구매자 주문]\n")
	for _, line := range sortedBreakdown(result.SourceProductBreakdown) {
		sb.WriteString(fmt.Sprintf("  %s\n", line))
	}
	sb.WriteString("\n")

	sb.WriteString("[상품 종류별 수량 - 출고 지시]\n")
	for _, line := range sortedBreakdown(result.TargetProductBreakdown) {
		sb.WriteString(fmt.Sprintf("  %s\n", line))
	}
	sb.WriteString("\n")

	sb.WriteString("[검증]\n")
	if result.SourceValidCount == result.TargetOrderCount {
		sb.WriteString(fmt.Sprintf("  ✓ 주문 건수 일치 (%d건)\n", result.SourceValidCount))
	}
	if result.SourceTotalQuantity == result.TargetTotalQuantity {
		sb.WriteString(fmt.Sprintf("  ✓ 상품 수량 일치 (%d개)\n", result.SourceTotalQuantity))
	}

	if len(result.Errors) > 0 {
		sb.WriteString("\n[오류]\n")
		for _, e := range result.Errors {
			sb.WriteString(fmt.Sprintf("  ✗ %s\n", e))
		}
	}

	if len(result.Warnings) > 0 {
		sb.WriteString("\n[경고]\n")
		for _, w := range result.Warnings {
			sb.WriteString(fmt.Sprintf("  ! %s\n", w))
		}
	}

	if len(result.Errors) == 0 {
		sb.WriteString("\n모든 검증을 통과했습니다.\n")
	}

	return sb.String()
}

func sortedBreakdown(breakdown map[string]int) []string {
	var keys []string
	for k := range breakdown {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var lines []string
	for _, k := range keys {
		lines = append(lines, fmt.Sprintf("%s: %d개", k, breakdown[k]))
	}
	return lines
}
