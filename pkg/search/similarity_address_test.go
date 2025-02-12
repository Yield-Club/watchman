package search

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAddress_findCountryCode(t *testing.T) {
	require.Equal(t, "US", findCountryCode("US"))
	require.Equal(t, "US", findCountryCode("USA"))
	require.Equal(t, "US", findCountryCode("UNITED STATES"))
	require.Equal(t, "US", findCountryCode("united states of america"))
}

func TestCompareAddress(t *testing.T) {
	var buf bytes.Buffer

	tests := []struct {
		name     string
		query    Address
		index    Address
		expected float64
	}{
		{
			name: "only_line1_exact",
			query: Address{
				Line1: "123 Main St",
			},
			index: Address{
				Line1: "123 Main St",
			},
			expected: 1.0,
		},
		{
			name: "only_line1_close",
			query: Address{
				Line1: "123 Main Street",
			},
			index: Address{
				Line1: "123 Main St",
			},
			expected: 0.941, // High but not exact due to Street vs St
		},
		{
			name: "only_line1_different_number",
			query: Address{
				Line1: "124 Main St", // Different building number
			},
			index: Address{
				Line1: "123 Main St",
			},
			expected: 0.941,
		},
		{
			name: "only_city",
			query: Address{
				City: "New York",
			},
			index: Address{
				City: "New York",
			},
			expected: 1.0,
		},
		{
			name: "similar_cities",
			query: Address{
				City: "Los Angeles",
			},
			index: Address{
				City: "Los Angles", // Common misspelling
			},
			expected: 0.964,
		},
		{
			name: "only_postal",
			query: Address{
				PostalCode: "90210",
			},
			index: Address{
				PostalCode: "90210",
			},
			expected: 1.0,
		},
		{
			name: "similar_addresses_different_units",
			query: Address{
				Line1: "123 Main St Apt 4B",
				City:  "New York",
				State: "NY",
			},
			index: Address{
				Line1: "123 Main St Apt 4C", // Different unit
				City:  "New York",
				State: "NY",
			},
			expected: 0.969,
		},
		{
			name: "similar_addresses_different_line2",
			query: Address{
				Line1: "123 Main St",
				Line2: "Apt 4B",
				City:  "New York",
				State: "NY",
			},
			index: Address{
				Line1: "123 Main St",
				Line2: "Apt 4C", // Different unit
				City:  "New York",
				State: "NY",
			},
			expected: 0.974,
		},
		{
			name: "country_code_vs_name",
			query: Address{
				Country: "United States",
			},
			index: Address{
				Country: "US",
			},
			expected: 1.0,
		},
		{
			name: "complex_partial_match",
			query: Address{
				Line1:      "1234 Broadway Suite 500",
				City:       "New York",
				State:      "NY",
				PostalCode: "10013",
				Country:    "US",
			},
			index: Address{
				Line1:      "1234 Broadway",
				City:       "New York",
				PostalCode: "10013",
			},
			expected: 0.792,
		},
		{
			name: "tricky_similar_but_different",
			query: Address{
				Line1: "45 Park Avenue South",
				City:  "New York",
				State: "NY",
			},
			index: Address{
				Line1: "45 Park Avenue North", // Different direction
				City:  "New York",
				State: "NY",
			},
			expected: 0.969,
		},
		{
			name: "ambiguous_addresses",
			query: Address{
				Line1: "100 Washington St",
				City:  "Boston",
			},
			index: Address{
				Line1: "100 Washington Ave", // Different street type
				City:  "Boston",
			},
			expected: 0.815,
		},
		{
			name: "missing_fields_comparison",
			query: Address{
				Line1: "555 Market St",
				City:  "San Francisco",
			},
			index: Address{
				Line1:   "555 Market St",
				Line2:   "Floor 2",
				City:    "San Francisco",
				State:   "CA",
				Country: "US",
			},
			expected: 1.0, // Should still be high as all provided fields match
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := compareAddress(&buf, tt.query, tt.index)
			require.InDelta(t, tt.expected, score, 0.001, "addresses should have expected similarity score (got %.2f, want %.2f)", score, tt.expected)
		})
	}
}

func TestCompareAddressesNoMatch(t *testing.T) {
	var buf bytes.Buffer

	tests := []struct {
		name     string
		query    Address
		index    Address
		expected float64
	}{
		{
			name: "completely_different_addresses",
			query: Address{
				Line1: "123 Main St",
				City:  "Boston",
				State: "MA",
			},
			index: Address{
				Line1: "456 Oak Ave",
				City:  "Chicago",
				State: "IL",
			},
			expected: 0.239,
		},
		{
			name: "similar_looking_but_different",
			query: Address{
				Line1: "1 World Trade Center",
				City:  "New York",
				State: "NY",
			},
			index: Address{
				Line1: "2 World Trade Center", // Different building
				City:  "New York",
				State: "NY",
			},
			expected: 0.886,
		},
		{
			name: "transposed_numbers",
			query: Address{
				Line1: "123 Main St",
				City:  "Anytown",
			},
			index: Address{
				Line1: "321 Main St",
				City:  "Anytown",
			},
			expected: 0.918,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := compareAddress(&buf, tt.query, tt.index)
			require.InDelta(t, tt.expected, score, 0.001, "different addresses should have low similarity score: %.2f", score)
		})
	}
}
