package helpers

import (
	"testing"
)

// ===============================================
// Tests for validation.go
// ===============================================

func TestValidateRequired(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		fieldName string
		wantError bool
	}{
		{"Empty string should error", "", "name", true},
		{"Whitespace only should error", "   ", "name", true},
		{"Valid value should pass", "John", "name", false},
		{"Value with spaces should pass", "John Doe", "name", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRequired(tt.value, tt.fieldName)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateRequired() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestValidateMaxLength(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		fieldName string
		maxLen    int
		wantError bool
	}{
		{"Empty string should pass", "", "name", 10, false},
		{"Within limit should pass", "Hello", "name", 10, false},
		{"At limit should pass", "0123456789", "name", 10, false},
		{"Over limit should error", "01234567890", "name", 10, true},
		{"Long string should error", "This is a very long string", "name", 10, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMaxLength(tt.value, tt.fieldName, tt.maxLen)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateMaxLength() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestValidateMinLength(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		fieldName string
		minLen    int
		wantError bool
	}{
		{"Empty string should error", "", "password", 6, true},
		{"Too short should error", "abc", "password", 6, true},
		{"At minimum should pass", "abcdef", "password", 6, false},
		{"Above minimum should pass", "abcdefghij", "password", 6, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMinLength(tt.value, tt.fieldName, tt.minLen)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateMinLength() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestValidatePositiveNumber(t *testing.T) {
	tests := []struct {
		name      string
		value     float64
		fieldName string
		wantError bool
	}{
		{"Zero should error", 0, "amount", true},
		{"Negative should error", -10.5, "amount", true},
		{"Positive should pass", 100.50, "amount", false},
		{"Small positive should pass", 0.01, "amount", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePositiveNumber(tt.value, tt.fieldName)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidatePositiveNumber() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestValidateRange(t *testing.T) {
	tests := []struct {
		name      string
		value     int
		min       int
		max       int
		fieldName string
		wantError bool
	}{
		{"Below range should error", 1999, 2000, 2100, "year", true},
		{"Above range should error", 2101, 2000, 2100, "year", true},
		{"At minimum should pass", 2000, 2000, 2100, "year", false},
		{"At maximum should pass", 2100, 2000, 2100, "year", false},
		{"Within range should pass", 2024, 2000, 2100, "year", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRange(tt.value, tt.min, tt.max, tt.fieldName)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateRange() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name      string
		email     string
		wantError bool
	}{
		{"Valid email should pass", "user@example.com", false},
		{"Valid email with subdomain", "user@mail.example.com", false},
		{"Valid email with plus", "user+tag@example.com", false},
		{"Invalid without @", "userexample.com", true},
		{"Invalid without domain", "user@", true},
		{"Invalid without user", "@example.com", true},
		{"Empty should error", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEmail(tt.email)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateEmail() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestValidateHexColor(t *testing.T) {
	tests := []struct {
		name      string
		color     string
		wantError bool
	}{
		{"Valid 6-char hex should pass", "#FF5733", false},
		{"Valid lowercase hex should pass", "#ff5733", false},
		{"Valid mixed case should pass", "#Ff5733", false},
		{"Missing hash should error", "FF5733", true},
		{"Invalid chars should error", "#GGGGGG", true},
		{"Too short should error", "#FFF", true},
		{"Too long should error", "#FF5733FF", true},
		{"Empty should pass (optional)", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateHexColor(tt.color)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateHexColor() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestValidateMonth(t *testing.T) {
	tests := []struct {
		name      string
		month     string
		wantError bool
	}{
		{"Janeiro should pass", "Janeiro", false},
		{"Fevereiro should pass", "Fevereiro", false},
		{"Dezembro should pass", "Dezembro", false},
		{"Março should pass", "Março", false},
		{"Acumulado should pass", "Acumulado", false},
		{"English January should error", "January", true},
		{"Invalid month should error", "InvalidMonth", true},
		{"Empty should error", "", true},
		{"Lowercase janeiro should error (case-sensitive)", "janeiro", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMonth(tt.month)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateMonth() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestValidateStatus(t *testing.T) {
	tests := []struct {
		name      string
		status    string
		wantError bool
	}{
		{"PAGO should pass", "PAGO", false},
		{"ABERTO should pass", "ABERTO", false},
		{"Lowercase pago should pass", "pago", false},
		{"Lowercase aberto should pass", "aberto", false},
		{"PENDING should error", "PENDING", true},
		{"Empty should error", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStatus(tt.status)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateStatus() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestValidateDayOfMonth(t *testing.T) {
	tests := []struct {
		name      string
		day       int
		wantError bool
	}{
		{"Day 1 should pass", 1, false},
		{"Day 15 should pass", 15, false},
		{"Day 31 should pass", 31, false},
		{"Day 0 should error", 0, true},
		{"Day 32 should error", 32, true},
		{"Negative day should error", -1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDayOfMonth(tt.day)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateDayOfMonth() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestCollectErrors(t *testing.T) {
	err1 := &ValidationError{Field: "name", Message: "required"}
	err2 := &ValidationError{Field: "email", Message: "invalid"}

	tests := []struct {
		name     string
		errors   []*ValidationError
		expected int
	}{
		{"No errors", []*ValidationError{nil, nil}, 0},
		{"One error", []*ValidationError{err1, nil}, 1},
		{"Two errors", []*ValidationError{err1, err2}, 2},
		{"All nil", []*ValidationError{nil, nil, nil}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CollectErrors(tt.errors...)
			if len(result) != tt.expected {
				t.Errorf("CollectErrors() = %v errors, want %v", len(result), tt.expected)
			}
		})
	}
}

func TestHasErrors(t *testing.T) {
	tests := []struct {
		name     string
		errors   []ValidationError
		expected bool
	}{
		{"Empty slice", []ValidationError{}, false},
		{"With error", []ValidationError{{Field: "test", Message: "error"}}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HasErrors(tt.errors)
			if result != tt.expected {
				t.Errorf("HasErrors() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// ===============================================
// Tests for sanitization.go
// ===============================================

func TestSanitizeString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Empty string", "", ""},
		{"Plain text", "Hello World", "Hello World"},
		{"Script tag", "<script>alert('XSS')</script>", ""},
		{"HTML tags", "<b>Bold</b> text", "Bold text"},
		{"Multiple spaces", "Hello    World", "Hello World"},
		{"Trim spaces", "  Hello  ", "Hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeString(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeString() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestEscapeHTML(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Empty string", "", ""},
		{"Plain text", "Hello", "Hello"},
		{"Script tag", "<script>", "&lt;script&gt;"},
		{"Ampersand", "Tom & Jerry", "Tom &amp; Jerry"},
		{"Quotes", `"quoted"`, "&#34;quoted&#34;"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EscapeHTML(tt.input)
			if result != tt.expected {
				t.Errorf("EscapeHTML() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestValidateNoScriptTags(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		wantError bool
	}{
		{"Plain text should pass", "Hello World", false},
		{"Script tag should error", "<script>alert(1)</script>", true},
		{"JavaScript URL should error", "javascript:alert(1)", true},
		{"Event handler should error", "onerror=alert(1)", true},
		{"Onclick should error", "onclick=doSomething()", true},
		{"Empty should pass", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNoScriptTags(tt.value, "field")
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateNoScriptTags() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestValidateMongoInjection(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		wantError bool
	}{
		{"Plain text should pass", "Hello World", false},
		{"$ne operator should error", "$ne", true},
		{"$gt operator should error", "$gt", true},
		{"$where operator should error", "$where", true},
		{"Normal text with dollar should pass", "Price: 100$", false},
		{"Empty should pass", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMongoInjection(tt.value, "field")
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateMongoInjection() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestValidatePathTraversal(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		wantError bool
	}{
		{"Plain text should pass", "file.txt", false},
		{"Parent directory should error", "../secret", true},
		{"Windows path traversal should error", "..\\secret", true},
		{"Encoded traversal should error", "%2e%2e/secret", true},
		{"Empty should pass", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePathTraversal(tt.value, "field")
			if (err != nil) != tt.wantError {
				t.Errorf("ValidatePathTraversal() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestSanitizeAlphanumeric(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Empty string", "", ""},
		{"Plain text", "Hello World", "Hello World"},
		{"With special chars", "Hello@World!", "HelloWorld"},
		{"With allowed punctuation", "John-Doe_2024", "John-Doe_2024"},
		{"With accents", "José María", "José María"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeAlphanumeric(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeAlphanumeric() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// Benchmarks
func BenchmarkValidateEmail(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ValidateEmail("user@example.com")
	}
}

func BenchmarkSanitizeString(b *testing.B) {
	input := "<script>alert('XSS')</script>Hello World"
	for i := 0; i < b.N; i++ {
		SanitizeString(input)
	}
}

func BenchmarkValidateNoScriptTags(b *testing.B) {
	input := "<script>alert(1)</script>"
	for i := 0; i < b.N; i++ {
		ValidateNoScriptTags(input, "field")
	}
}
