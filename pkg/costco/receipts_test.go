package costco

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReceiptItem_IsDiscount(t *testing.T) {
	tests := []struct {
		name     string
		item     ReceiptItem
		expected bool
	}{
		{
			name: "discount item with slash prefix",
			item: ReceiptItem{
				ItemNumber:        "363064",
				ItemDescription01: "/1553261",
				Amount:            -4.00,
				Unit:              -1,
			},
			expected: true,
		},
		{
			name: "regular item with positive amount",
			item: ReceiptItem{
				ItemNumber:        "1553261",
				ItemDescription01: "GUAC BOWL",
				Amount:            13.99,
				Unit:              1,
			},
			expected: false,
		},
		{
			name: "return item with negative amount but no slash",
			item: ReceiptItem{
				ItemNumber:        "1469292",
				ItemDescription01: "RED GRAPE",
				Amount:            -7.49,
				Unit:              -1,
			},
			expected: false,
		},
		{
			name: "negative amount but positive unit",
			item: ReceiptItem{
				ItemNumber:        "123456",
				ItemDescription01: "/1111111",
				Amount:            -5.00,
				Unit:              1, // Positive unit, shouldn't be a discount
			},
			expected: false,
		},
		{
			name: "slash prefix but positive amount",
			item: ReceiptItem{
				ItemNumber:        "123456",
				ItemDescription01: "/1111111",
				Amount:            5.00, // Positive amount
				Unit:              -1,
			},
			expected: false,
		},
		{
			name: "empty description",
			item: ReceiptItem{
				ItemNumber:        "123456",
				ItemDescription01: "",
				Amount:            -5.00,
				Unit:              -1,
			},
			expected: false,
		},
		{
			name: "description with slash but not at start",
			item: ReceiptItem{
				ItemNumber:        "123456",
				ItemDescription01: "ITEM/1111111",
				Amount:            -5.00,
				Unit:              -1,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.item.IsDiscount()
			assert.Equal(t, tt.expected, result, "IsDiscount() should return %v for %s", tt.expected, tt.name)
		})
	}
}

func TestReceiptItem_GetParentItemNumber(t *testing.T) {
	tests := []struct {
		name     string
		item     ReceiptItem
		expected string
	}{
		{
			name: "discount item returns parent item number",
			item: ReceiptItem{
				ItemNumber:        "363064",
				ItemDescription01: "/1553261",
				Amount:            -4.00,
				Unit:              -1,
			},
			expected: "1553261",
		},
		{
			name: "regular item returns empty string",
			item: ReceiptItem{
				ItemNumber:        "1553261",
				ItemDescription01: "GUAC BOWL",
				Amount:            13.99,
				Unit:              1,
			},
			expected: "",
		},
		{
			name: "return item returns empty string",
			item: ReceiptItem{
				ItemNumber:        "1469292",
				ItemDescription01: "RED GRAPE",
				Amount:            -7.49,
				Unit:              -1,
			},
			expected: "",
		},
		{
			name: "discount with different item number format",
			item: ReceiptItem{
				ItemNumber:        "999999",
				ItemDescription01: "/12345",
				Amount:            -2.50,
				Unit:              -1,
			},
			expected: "12345",
		},
		{
			name: "item with slash in middle of description",
			item: ReceiptItem{
				ItemNumber:        "123456",
				ItemDescription01: "ITEM/CODE",
				Amount:            -5.00,
				Unit:              -1,
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.item.GetParentItemNumber()
			assert.Equal(t, tt.expected, result, "GetParentItemNumber() should return '%s' for %s", tt.expected, tt.name)
		})
	}
}

func TestReceiptItem_RealWorldDiscountExample(t *testing.T) {
	// Real data from user's Costco receipt
	guacItem := ReceiptItem{
		ItemNumber:        "1553261",
		ItemDescription01: "GUAC BOWL",
		Amount:            13.99,
		Unit:              1,
	}

	discountItem := ReceiptItem{
		ItemNumber:        "363064",
		ItemDescription01: "/1553261",
		Amount:            -4.00,
		Unit:              -1,
	}

	// Guac item should not be identified as a discount
	assert.False(t, guacItem.IsDiscount(), "Regular guac item should not be a discount")
	assert.Equal(t, "", guacItem.GetParentItemNumber(), "Regular item should have empty parent number")

	// Discount item should be identified correctly
	assert.True(t, discountItem.IsDiscount(), "Discount item should be identified as a discount")
	assert.Equal(t, "1553261", discountItem.GetParentItemNumber(), "Discount should reference guac item number")

	// Verify we can calculate net amount
	netAmount := guacItem.Amount + discountItem.Amount
	assert.Equal(t, 9.99, netAmount, "Net amount should be 13.99 - 4.00 = 9.99")
}

func TestReceiptItem_RealWorldReturnExample(t *testing.T) {
	// Real data from user's Costco return receipt
	returnItem := ReceiptItem{
		ItemNumber:        "1469292",
		ItemDescription01: "RED GRAPE",
		Amount:            -7.49,
		Unit:              -1,
	}

	// Return items should NOT be identified as discounts
	assert.False(t, returnItem.IsDiscount(), "Return item should not be identified as a discount")
	assert.Equal(t, "", returnItem.GetParentItemNumber(), "Return item should have empty parent number")
}

func TestReceiptItem_ProcessingWorkflow(t *testing.T) {
	// Simulate a receipt with multiple items including a discount
	// SubTotal should be: 13.99 (guac) + -4.00 (discount) + 6.99 (broccoli) = 16.98
	receipt := Receipt{
		SubTotal:       16.98,
		InstantSavings: 4.00,
		ItemArray: []ReceiptItem{
			{
				ItemNumber:        "1553261",
				ItemDescription01: "GUAC BOWL",
				Amount:            13.99,
				Unit:              1,
			},
			{
				ItemNumber:        "363064",
				ItemDescription01: "/1553261",
				Amount:            -4.00,
				Unit:              -1,
			},
			{
				ItemNumber:        "5623",
				ItemDescription01: "BROCCOLI",
				Amount:            6.99,
				Unit:              1,
			},
		},
	}

	// Test filtering out discounts
	var regularItems []ReceiptItem
	for _, item := range receipt.ItemArray {
		if !item.IsDiscount() {
			regularItems = append(regularItems, item)
		}
	}

	assert.Len(t, regularItems, 2, "Should have 2 regular items after filtering discounts")
	assert.Equal(t, "GUAC BOWL", regularItems[0].ItemDescription01)
	assert.Equal(t, "BROCCOLI", regularItems[1].ItemDescription01)

	// Test building net amounts map
	itemAmounts := make(map[string]float64)
	for _, item := range receipt.ItemArray {
		if item.IsDiscount() {
			parentNum := item.GetParentItemNumber()
			itemAmounts[parentNum] += item.Amount
		} else {
			itemAmounts[item.ItemNumber] += item.Amount
		}
	}

	assert.Equal(t, 9.99, itemAmounts["1553261"], "Guac should have net amount of 9.99")
	assert.Equal(t, 6.99, itemAmounts["5623"], "Broccoli should have original amount of 6.99")

	// Verify total matches
	var total float64
	for _, amount := range itemAmounts {
		total += amount
	}
	assert.Equal(t, receipt.SubTotal, total, "Sum of net amounts should equal receipt subtotal")
}
