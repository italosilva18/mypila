package handlers

import (
	"testing"
)

func TestParsePage(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"Valid page 1", "1", 1},
		{"Valid page 5", "5", 5},
		{"Valid page 100", "100", 100},
		{"Zero page", "0", 1},
		{"Negative page", "-5", 1},
		{"Invalid string", "abc", 1},
		{"Empty string", "", 1},
		{"Float number", "3.14", 1},
		{"Large number", "999999", 999999},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parsePage(tt.input)
			if result != tt.expected {
				t.Errorf("parsePage(%s) = %d; want %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseLimit(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"Default limit", "50", 50},
		{"Valid limit 1", "1", 1},
		{"Valid limit 100", "100", 100},
		{"Exceeds max (101)", "101", 100},
		{"Exceeds max (200)", "200", 100},
		{"Exceeds max (1000)", "1000", 100},
		{"Zero limit", "0", 50},
		{"Negative limit", "-10", 50},
		{"Invalid string", "xyz", 50},
		{"Empty string", "", 50},
		{"Float number", "25.5", 50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseLimit(tt.input)
			if result != tt.expected {
				t.Errorf("parseLimit(%s) = %d; want %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseLimitEnforcement(t *testing.T) {
	// Test that limit never exceeds 100
	limits := []string{"101", "150", "200", "500", "1000", "999999"}

	for _, limitStr := range limits {
		result := parseLimit(limitStr)
		if result > 100 {
			t.Errorf("parseLimit(%s) = %d; exceeds maximum of 100", limitStr, result)
		}
	}
}

func TestParsePageNeverNegative(t *testing.T) {
	// Test that page is never negative or zero
	pages := []string{"-1", "-10", "-100", "0"}

	for _, pageStr := range pages {
		result := parsePage(pageStr)
		if result < 1 {
			t.Errorf("parsePage(%s) = %d; should be at least 1", pageStr, result)
		}
	}
}

func TestPaginationDefaults(t *testing.T) {
	// Test default values
	t.Run("Default page", func(t *testing.T) {
		result := parsePage("")
		if result != 1 {
			t.Errorf("Default page should be 1, got %d", result)
		}
	})

	t.Run("Default limit", func(t *testing.T) {
		result := parseLimit("")
		if result != 50 {
			t.Errorf("Default limit should be 50, got %d", result)
		}
	})
}

func TestPaginationBoundaries(t *testing.T) {
	t.Run("Minimum valid page", func(t *testing.T) {
		result := parsePage("1")
		if result != 1 {
			t.Errorf("Minimum page should be 1, got %d", result)
		}
	})

	t.Run("Maximum valid limit", func(t *testing.T) {
		result := parseLimit("100")
		if result != 100 {
			t.Errorf("Maximum limit should be 100, got %d", result)
		}
	})

	t.Run("Minimum valid limit", func(t *testing.T) {
		result := parseLimit("1")
		if result != 1 {
			t.Errorf("Minimum limit should be 1, got %d", result)
		}
	})
}

// Benchmark tests
func BenchmarkParsePage(b *testing.B) {
	for i := 0; i < b.N; i++ {
		parsePage("42")
	}
}

func BenchmarkParseLimit(b *testing.B) {
	for i := 0; i < b.N; i++ {
		parseLimit("50")
	}
}

func BenchmarkParsePageInvalid(b *testing.B) {
	for i := 0; i < b.N; i++ {
		parsePage("invalid")
	}
}

func BenchmarkParseLimitInvalid(b *testing.B) {
	for i := 0; i < b.N; i++ {
		parseLimit("invalid")
	}
}
