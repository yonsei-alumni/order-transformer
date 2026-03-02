package main

// Transform converts source orders to shipping orders, excluding cancelled orders.
func Transform(orders []SourceOrder) []ShippingOrder {
	var result []ShippingOrder
	for _, order := range orders {
		if order.IsCancelled() {
			continue
		}

		shipping := ShippingOrder{
			OrderNumber:     order.OrderNumber,
			RecipientName:   order.RecipientName,
			FullAddress:     order.CombinedAddress(),
			PostalCode:      order.PostalCode,
			Phone:           order.RecipientPhone,
			Quantity:        order.Quantity,
			DeliveryMessage: order.DeliveryMemo,
			ProductName:     order.CombinedProductName(),
		}

		result = append(result, shipping)
	}
	return result
}
