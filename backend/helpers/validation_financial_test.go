package helpers

import (
	"math"
	"strings"
	"testing"
)

func TestValidateAmount(t *testing.T) {
	tests := []struct {
		name      string
		value     float64
		fieldName string
		wantErr   bool
		errMsg    string
	}{
		{"Valid amount", 100.50, "amount", false, ""},
		{"Valid minimum amount", 0.01, "amount", false, ""},
		{"Valid large amount", 999999999.99, "amount", false, ""},
		{"Zero should error", 0, "amount", true, "deve ser maior que zero"},
		{"Negative should error", -100, "amount", true, "deve ser maior que zero"},
		{"Exceeds maximum should error", 1000000000.00, "amount", true, "excede o valor maximo"},
		{"Too many decimals should error", 100.123, "amount", true, "maximo 2 casas decimais"},
		{"NaN should error", math.NaN(), "amount", true, "valor numerico invalido"},
		{"Positive infinity should error", math.Inf(1), "amount", true, "valor numerico invalido"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAmount(tt.value, tt.fieldName)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateAmount() expected error, got nil")
				} else if !strings.Contains(err.Message, tt.errMsg) {
					t.Errorf("ValidateAmount() error = %v, want contains %v", err.Message, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateAmount() unexpected error: %v", err.Message)
				}
			}
		})
	}
}

func TestValidateAmountAllowZero(t *testing.T) {
	tests := []struct {
		name      string
		value     float64
		fieldName string
		wantErr   bool
		errMsg    string
	}{
		{"Valid amount", 100.50, "budget", false, ""},
		{"Zero should pass", 0, "budget", false, ""},
		{"Negative should error", -100, "budget", true, "maior ou igual a zero"},
		{"Exceeds max should error", 1000000000.00, "budget", true, "excede o valor maximo"},
		{"NaN should error", math.NaN(), "budget", true, "valor numerico invalido"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAmountAllowZero(tt.value, tt.fieldName)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateAmountAllowZero() expected error, got nil")
				} else if !strings.Contains(err.Message, tt.errMsg) {
					t.Errorf("ValidateAmountAllowZero() error = %v, want contains %v", err.Message, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateAmountAllowZero() unexpected error: %v", err.Message)
				}
			}
		})
	}
}

func TestValidateDiscountValue(t *testing.T) {
	tests := []struct {
		name         string
		value        float64
		discountType string
		fieldName    string
		wantErr      bool
		errMsg       string
	}{
		{"Valid percent 0", 0, "PERCENT", "discount", false, ""},
		{"Valid percent 50", 50, "PERCENT", "discount", false, ""},
		{"Valid percent 100", 100, "PERCENT", "discount", false, ""},
		{"Percent over 100", 101, "PERCENT", "discount", true, "entre 0 e 100"},
		{"Valid value discount", 500.00, "VALUE", "discount", false, ""},
		{"Negative percent", -10, "PERCENT", "discount", true, "maior ou igual a zero"},
		{"Negative value", -100, "VALUE", "discount", true, "maior ou igual a zero"},
		{"NaN should error", math.NaN(), "VALUE", "discount", true, "valor numerico invalido"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDiscountValue(tt.value, tt.discountType, tt.fieldName)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateDiscountValue() expected error, got nil")
				} else if !strings.Contains(err.Message, tt.errMsg) {
					t.Errorf("ValidateDiscountValue() error = %v, want contains %v", err.Message, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateDiscountValue() unexpected error: %v", err.Message)
				}
			}
		})
	}
}

func TestValidateQuantity(t *testing.T) {
	tests := []struct {
		name      string
		value     float64
		fieldName string
		wantErr   bool
		errMsg    string
	}{
		{"Valid integer", 10, "quantity", false, ""},
		{"Valid fractional", 0.5, "quantity", false, ""},
		{"Zero should error", 0, "quantity", true, "maior que zero"},
		{"Negative should error", -5, "quantity", true, "maior que zero"},
		{"Exceeds max", 1000000, "quantity", true, "excede o valor maximo"},
		{"NaN should error", math.NaN(), "quantity", true, "valor numerico invalido"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateQuantity(tt.value, tt.fieldName)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateQuantity() expected error, got nil")
				} else if !strings.Contains(err.Message, tt.errMsg) {
					t.Errorf("ValidateQuantity() error = %v, want contains %v", err.Message, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateQuantity() unexpected error: %v", err.Message)
				}
			}
		})
	}
}

func TestValidateBudget(t *testing.T) {
	tests := []struct {
		name      string
		value     float64
		fieldName string
		wantErr   bool
	}{
		{"Valid budget", 5000.00, "budget", false},
		{"Zero budget", 0, "budget", false},
		{"Negative budget", -100, "budget", true},
		{"Over maximum", 1000000000.00, "budget", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateBudget(tt.value, tt.fieldName)
			if tt.wantErr && err == nil {
				t.Errorf("ValidateBudget() expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("ValidateBudget() unexpected error: %v", err.Message)
			}
		})
	}
}

func TestRoundToTwoDecimals(t *testing.T) {
	tests := []struct {
		name     string
		value    float64
		expected float64
	}{
		{"Already 2 decimals", 100.50, 100.50},
		{"Round down", 100.504, 100.50},
		{"Round up", 100.506, 100.51},
		{"No decimals", 100, 100.00},
		{"Zero", 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RoundToTwoDecimals(tt.value)
			if math.Abs(result-tt.expected) > 0.001 {
				t.Errorf("RoundToTwoDecimals(%v) = %v, want %v", tt.value, result, tt.expected)
			}
		})
	}
}

func TestIsValidFiniteNumber(t *testing.T) {
	tests := []struct {
		name     string
		value    float64
		expected bool
	}{
		{"Normal number", 100.50, true},
		{"Zero", 0, true},
		{"Negative", -100, true},
		{"NaN", math.NaN(), false},
		{"Positive infinity", math.Inf(1), false},
		{"Negative infinity", math.Inf(-1), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidFiniteNumber(tt.value)
			if result != tt.expected {
				t.Errorf("isValidFiniteNumber(%v) = %v, want %v", tt.value, result, tt.expected)
			}
		})
	}
}
