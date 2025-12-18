package helpers

import (
	"html"
	"regexp"
	"strings"

	"github.com/microcosm-cc/bluemonday"
)

// ===============================================
// SECURITY: XSS Prevention & HTML Sanitization
// ===============================================

// SanitizeString removes all HTML tags and dangerous characters from user input
// Use this for fields that should NOT contain any HTML (names, simple descriptions)
func SanitizeString(input string) string {
	if input == "" {
		return input
	}

	// Create a strict policy that strips ALL HTML
	policy := bluemonday.StrictPolicy()

	// Sanitize the input
	sanitized := policy.Sanitize(input)

	// Additional cleanup: remove any remaining dangerous characters
	sanitized = strings.TrimSpace(sanitized)

	// Normalize whitespace
	sanitized = normalizeWhitespace(sanitized)

	return sanitized
}

// SanitizeHTML allows only safe HTML tags for rich text content
// Use this when you need to preserve some formatting (descriptions with basic formatting)
func SanitizeHTML(input string) string {
	if input == "" {
		return input
	}

	// Create a policy that allows only safe HTML tags
	policy := bluemonday.UGCPolicy()

	// Further restrict to only basic formatting
	policy = policy.AllowElements("p", "br", "strong", "em", "u")

	// Sanitize the input
	sanitized := policy.Sanitize(input)

	return strings.TrimSpace(sanitized)
}

// EscapeHTML escapes special HTML characters to prevent XSS
// This is the safest option - converts <script> to &lt;script&gt;
func EscapeHTML(input string) string {
	if input == "" {
		return input
	}

	return html.EscapeString(input)
}

// SanitizeAlphanumeric allows only alphanumeric characters, spaces, and basic punctuation
// Use for names, titles, etc.
func SanitizeAlphanumeric(input string) string {
	if input == "" {
		return input
	}

	// Allow letters (including unicode), numbers, spaces, and basic punctuation
	reg := regexp.MustCompile(`[^\p{L}\p{N}\s\.\-_,]`)
	sanitized := reg.ReplaceAllString(input, "")

	return strings.TrimSpace(sanitized)
}

// ValidateSQLInjection checks for common SQL injection patterns
// Returns error if suspicious patterns are detected
func ValidateSQLInjection(value, fieldName string) *ValidationError {
	if value == "" {
		return nil
	}

	// Common SQL injection patterns
	sqlPatterns := []string{
		`(?i)\b(SELECT|INSERT|UPDATE|DELETE|DROP|CREATE|ALTER|EXEC|EXECUTE)\b`,
		`(?i)\b(UNION|OR|AND)\b.*=.*`,
		`--`,
		`/\*|\*/`,
		`;.*\b(SELECT|INSERT|UPDATE|DELETE)\b`,
		`\bxp_\w+`,
		`\bsp_\w+`,
	}

	for _, pattern := range sqlPatterns {
		matched, _ := regexp.MatchString(pattern, value)
		if matched {
			return &ValidationError{
				Field:   fieldName,
				Message: "Conteúdo contém caracteres não permitidos",
			}
		}
	}

	return nil
}

// ValidateNoScriptTags validates that input doesn't contain script tags
func ValidateNoScriptTags(value, fieldName string) *ValidationError {
	if value == "" {
		return nil
	}

	// Check for script tags (case insensitive)
	scriptPattern := `(?i)<script|javascript:|onerror=|onload=|onclick=`
	matched, _ := regexp.MatchString(scriptPattern, value)

	if matched {
		return &ValidationError{
			Field:   fieldName,
			Message: "Conteúdo contém código não permitido",
		}
	}

	return nil
}

// ValidatePathTraversal checks for path traversal attempts
func ValidatePathTraversal(value, fieldName string) *ValidationError {
	if value == "" {
		return nil
	}

	// Check for path traversal patterns
	pathPatterns := []string{
		`\.\.[\\/]`,
		`[\\/]\.\.`,
		`%2e%2e`,
		`%252e%252e`,
	}

	for _, pattern := range pathPatterns {
		matched, _ := regexp.MatchString(pattern, strings.ToLower(value))
		if matched {
			return &ValidationError{
				Field:   fieldName,
				Message: "Caminho inválido detectado",
			}
		}
	}

	return nil
}

// normalizeWhitespace replaces multiple whitespace characters with a single space
func normalizeWhitespace(input string) string {
	reg := regexp.MustCompile(`\s+`)
	return reg.ReplaceAllString(input, " ")
}

// ValidateMongoInjection prevents NoSQL injection by validating MongoDB operators
func ValidateMongoInjection(value, fieldName string) *ValidationError {
	if value == "" {
		return nil
	}

	// Check for MongoDB operators
	mongoPattern := `\$\w+`
	matched, _ := regexp.MatchString(mongoPattern, value)

	if matched {
		return &ValidationError{
			Field:   fieldName,
			Message: "Operadores não permitidos detectados",
		}
	}

	return nil
}
