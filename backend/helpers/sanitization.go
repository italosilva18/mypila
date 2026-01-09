package helpers

import (
	"html"
	"regexp"
	"strings"

	"github.com/microcosm-cc/bluemonday"
)

// Pre-compiled regex patterns and policies for better performance (singleton)
var (
	strictPolicy      = bluemonday.StrictPolicy()
	whitespaceRegex   = regexp.MustCompile(`\s+`)
	alphanumericRegex = regexp.MustCompile(`[^\p{L}\p{N}\s\.\-_,]`)
	scriptTagRegex    = regexp.MustCompile(`(?i)<script|javascript:|onerror=|onload=|onclick=`)
	mongoOperatorRegex = regexp.MustCompile(`\$\w+`)
)

// Pre-compiled SQL injection patterns
var sqlPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)\b(SELECT|INSERT|UPDATE|DELETE|DROP|CREATE|ALTER|EXEC|EXECUTE)\b`),
	regexp.MustCompile(`(?i)\b(UNION|OR|AND)\b.*=.*`),
	regexp.MustCompile(`--`),
	regexp.MustCompile(`/\*|\*/`),
	regexp.MustCompile(`;.*\b(SELECT|INSERT|UPDATE|DELETE)\b`),
	regexp.MustCompile(`\bxp_\w+`),
	regexp.MustCompile(`\bsp_\w+`),
}

// Pre-compiled path traversal patterns
var pathPatterns = []*regexp.Regexp{
	regexp.MustCompile(`\.\.[\\/]`),
	regexp.MustCompile(`[\\/]\.\.`),
	regexp.MustCompile(`%2e%2e`),
	regexp.MustCompile(`%252e%252e`),
}

// ===============================================
// SECURITY: XSS Prevention & HTML Sanitization
// ===============================================

// SanitizeString removes all HTML tags and dangerous characters from user input
// Use this for fields that should NOT contain any HTML (names, simple descriptions)
func SanitizeString(input string) string {
	if input == "" {
		return input
	}

	// Use pre-created strict policy (singleton)
	policy := strictPolicy

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

	// Use pre-compiled regex for better performance
	sanitized := alphanumericRegex.ReplaceAllString(input, "")

	return strings.TrimSpace(sanitized)
}

// ValidateSQLInjection checks for common SQL injection patterns
// Returns error if suspicious patterns are detected
func ValidateSQLInjection(value, fieldName string) *ValidationError {
	if value == "" {
		return nil
	}

	// Use pre-compiled regex patterns for better performance
	for _, pattern := range sqlPatterns {
		if pattern.MatchString(value) {
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

	// Use pre-compiled regex for better performance
	if scriptTagRegex.MatchString(value) {
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

	// Use pre-compiled regex patterns for better performance
	lowerValue := strings.ToLower(value)
	for _, pattern := range pathPatterns {
		if pattern.MatchString(lowerValue) {
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
	// Use pre-compiled regex for better performance
	return whitespaceRegex.ReplaceAllString(input, " ")
}

// ValidateMongoInjection prevents NoSQL injection by validating MongoDB operators
func ValidateMongoInjection(value, fieldName string) *ValidationError {
	if value == "" {
		return nil
	}

	// Use pre-compiled regex for better performance
	if mongoOperatorRegex.MatchString(value) {
		return &ValidationError{
			Field:   fieldName,
			Message: "Operadores não permitidos detectados",
		}
	}

	return nil
}
