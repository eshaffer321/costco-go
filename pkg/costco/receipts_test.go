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
		{
			name: "discount with whitespace after slash (real-world data)",
			item: ReceiptItem{
				ItemNumber:        "363581",
				ItemDescription01: "/ 1857091", // Space after slash - real Costco data
				Amount:            -2.90,
				Unit:              -1,
			},
			expected: "1857091", // Should trim whitespace
		},
		{
			name: "discount with multiple spaces after slash",
			item: ReceiptItem{
				ItemNumber:        "363582",
				ItemDescription01: "/  1569515", // Multiple spaces
				Amount:            -3.50,
				Unit:              -1,
			},
			expected: "1569515", // Should trim all whitespace
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

func TestNetDiscounts(t *testing.T) {
	t.Run("applies discount via item number match", func(t *testing.T) {
		items := []ReceiptItem{
			{ItemNumber: "1553261", ItemDescription01: "GUAC BOWL", Amount: 13.99, Unit: 1},
			{ItemNumber: "363064", ItemDescription01: "/1553261", Amount: -4.00, Unit: -1},
		}
		netted, orphaned := NetDiscounts(items)
		assert.Len(t, netted, 1)
		assert.Len(t, orphaned, 0)
		assert.Equal(t, 9.99, netted[0].Amount)
		assert.Equal(t, "GUAC BOWL", netted[0].ItemDescription01)
	})

	t.Run("applies discount via exact description match", func(t *testing.T) {
		// discount's ItemDescription01 is "/AAA BATTERY" → ref is "AAA BATTERY"
		// parent item description is exactly "AAA BATTERY"
		items := []ReceiptItem{
			{ItemNumber: "379938", ItemDescription01: "AAA BATTERY", Amount: 14.99, Unit: 1},
			{ItemNumber: "999001", ItemDescription01: "/AAA BATTERY", Amount: -2.50, Unit: -1},
		}
		netted, orphaned := NetDiscounts(items)
		assert.Len(t, netted, 1)
		assert.Len(t, orphaned, 0)
		assert.Equal(t, 12.49, netted[0].Amount)
	})

	t.Run("applies discount via partial description match", func(t *testing.T) {
		// Real receipts: parent item is "AA/AAA BATTERY" but coupon references "/AAA BATTERY".
		// Neither item-number nor exact-description lookup succeeds; substring fallback must fire.
		items := []ReceiptItem{
			{ItemNumber: "379938", ItemDescription01: "AA/AAA BATTERY", Amount: 14.99, Unit: 1},
			{ItemNumber: "999001", ItemDescription01: "/AAA BATTERY", Amount: -2.50, Unit: -1},
		}
		netted, orphaned := NetDiscounts(items)
		assert.Len(t, netted, 1, "should have 1 item after netting partial-description-referenced discount")
		assert.Len(t, orphaned, 0, "discount should be matched, not orphaned")
		assert.Equal(t, 12.49, netted[0].Amount, "discount should reduce price")
		assert.Equal(t, "AA/AAA BATTERY", netted[0].ItemDescription01)
	})

	t.Run("returns orphaned discount when no parent found", func(t *testing.T) {
		items := []ReceiptItem{
			{ItemNumber: "111", ItemDescription01: "CHICKEN", Amount: 10.00, Unit: 1},
			{ItemNumber: "999", ItemDescription01: "/UNKNOWN ITEM", Amount: -2.00, Unit: -1},
		}
		netted, orphaned := NetDiscounts(items)
		assert.Len(t, netted, 1)
		assert.Len(t, orphaned, 1)
		assert.Equal(t, 10.00, netted[0].Amount, "unmatched discount must not affect other items")
		assert.Equal(t, "/UNKNOWN ITEM", orphaned[0].ItemDescription01)
	})

	t.Run("multiple discounts on same item", func(t *testing.T) {
		items := []ReceiptItem{
			{ItemNumber: "1000", ItemDescription01: "TOILET PAPER", Amount: 25.99, Unit: 1},
			{ItemNumber: "2001", ItemDescription01: "/1000", Amount: -2.00, Unit: -1},
			{ItemNumber: "2002", ItemDescription01: "/1000", Amount: -1.00, Unit: -1},
		}
		netted, orphaned := NetDiscounts(items)
		assert.Len(t, netted, 1)
		assert.Len(t, orphaned, 0)
		assert.Equal(t, 22.99, netted[0].Amount)
	})

	t.Run("empty input returns empty slices", func(t *testing.T) {
		netted, orphaned := NetDiscounts(nil)
		assert.Empty(t, netted)
		assert.Empty(t, orphaned)
	})

	t.Run("no discounts returns all items unchanged", func(t *testing.T) {
		items := []ReceiptItem{
			{ItemNumber: "1", ItemDescription01: "MILK", Amount: 4.99, Unit: 1},
			{ItemNumber: "2", ItemDescription01: "EGGS", Amount: 6.99, Unit: 1},
		}
		netted, orphaned := NetDiscounts(items)
		assert.Len(t, netted, 2)
		assert.Len(t, orphaned, 0)
	})

	t.Run("applies discount via word overlap when item brand differs from coupon category", func(t *testing.T) {
		// Real receipt (barcode 21134301000462605131128):
		//   item_number=1627198  desc01="DURACELL AAA"   (the actual product)
		//   item_number=379938   desc01="/AAA BATTERY"   (the coupon)
		//
		// "DURACELL AAA" does not contain "AAA BATTERY", and "AAA BATTERY" does not
		// contain "DURACELL AAA", so neither the substring fallback nor exact match
		// works. The shared word "AAA" must be used to identify the parent.
		items := []ReceiptItem{
			{ItemNumber: "1627198", ItemDescription01: "DURACELL AAA", ItemDescription02: "40PK BATTERIES P432", Amount: 20.99, Unit: 1},
			{ItemNumber: "379938", ItemDescription01: "/AAA BATTERY", Amount: -2.50, Unit: -1},
		}
		netted, orphaned := NetDiscounts(items)
		assert.Len(t, netted, 1, "should have 1 item after netting word-overlap-matched discount")
		assert.Len(t, orphaned, 0, "discount should be matched via word overlap, not orphaned")
		assert.InDelta(t, 18.49, netted[0].Amount, 0.001, "discount should be applied to DURACELL AAA")
		assert.Equal(t, "DURACELL AAA", netted[0].ItemDescription01)
	})

	t.Run("word overlap picks best match when multiple candidates share words", func(t *testing.T) {
		// Two items where only word-overlap can distinguish which is the right parent.
		// "DURACELL AA"  → word "AA" is 2 chars, below the 3-char floor → score 0
		// "DURACELL AAA" → word "AAA" is 3 chars and appears in "AAA BATTERY" → score 1
		// Neither candidate is a substring of the ref or vice versa, so no ambiguity
		// from map-iteration order in the substring step.
		items := []ReceiptItem{
			{ItemNumber: "100", ItemDescription01: "DURACELL AA", Amount: 15.99, Unit: 1},  // score 0: "AA" < 3 chars
			{ItemNumber: "200", ItemDescription01: "DURACELL AAA", Amount: 20.99, Unit: 1}, // score 1: "AAA" matches
			{ItemNumber: "379938", ItemDescription01: "/AAA BATTERY", Amount: -2.50, Unit: -1},
		}
		netted, orphaned := NetDiscounts(items)
		assert.Len(t, netted, 2)
		assert.Len(t, orphaned, 0, "discount should be matched to best candidate")
		for _, item := range netted {
			if item.ItemDescription01 == "DURACELL AAA" {
				assert.InDelta(t, 18.49, item.Amount, 0.001, "discount should go to the higher word-overlap match")
			} else {
				assert.InDelta(t, 15.99, item.Amount, 0.001, "DURACELL AA should be unaffected (no word overlap)")
			}
		}
	})
}
