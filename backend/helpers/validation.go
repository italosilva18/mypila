package helpers

import (
	"fmt"
	"math"
	"regexp"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// Pre-compiled regex patterns for better performance
var (
	emailRegex    = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	hexColorRegex = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)
)

// ValidationError representa um erro de validação com mensagem em português
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationErrors representa múltiplos erros de validação
type ValidationErrors struct {
	Errors []ValidationError `json:"errors"`
}

// SendValidationError envia uma resposta 400 com erro de validação
func SendValidationError(c *fiber.Ctx, field, message string) error {
	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
		"error": message,
		"field": field,
	})
}

// SendValidationErrors envia múltiplos erros de validação
func SendValidationErrors(c *fiber.Ctx, errors []ValidationError) error {
	return c.Status(fiber.StatusBadRequest).JSON(ValidationErrors{
		Errors: errors,
	})
}

// ValidateRequired valida se um campo não está vazio
func ValidateRequired(value, fieldName string) *ValidationError {
	if strings.TrimSpace(value) == "" {
		return &ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("%s não pode ser vazio", fieldName),
		}
	}
	return nil
}

// ValidateMaxLength valida o comprimento máximo de uma string
func ValidateMaxLength(value, fieldName string, maxLength int) *ValidationError {
	if len(value) > maxLength {
		return &ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("%s deve ter no máximo %d caracteres", fieldName, maxLength),
		}
	}
	return nil
}

// ValidateMinLength valida o comprimento mínimo de uma string
func ValidateMinLength(value, fieldName string, minLength int) *ValidationError {
	if len(value) < minLength {
		return &ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("%s deve ter no mínimo %d caracteres", fieldName, minLength),
		}
	}
	return nil
}

// ValidatePositiveNumber valida se um número é maior que zero
func ValidatePositiveNumber(value float64, fieldName string) *ValidationError {
	if value <= 0 {
		return &ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("%s deve ser maior que zero", fieldName),
		}
	}
	return nil
}

// ValidateRange valida se um valor inteiro está dentro de um intervalo
func ValidateRange(value int, min, max int, fieldName string) *ValidationError {
	if value < min || value > max {
		return &ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("%s deve estar entre %d e %d", fieldName, min, max),
		}
	}
	return nil
}

// ValidateEmail valida o formato de um email
func ValidateEmail(email string) *ValidationError {
	if email == "" {
		return &ValidationError{
			Field:   "email",
			Message: "Email não pode ser vazio",
		}
	}

	// Use pre-compiled regex for better performance
	if !emailRegex.MatchString(email) {
		return &ValidationError{
			Field:   "email",
			Message: "Formato de email inválido",
		}
	}
	return nil
}

// ValidateHexColor valida se uma string é uma cor hexadecimal válida
func ValidateHexColor(color string) *ValidationError {
	if color == "" {
		return nil // Cor pode ser opcional
	}

	// Use pre-compiled regex for better performance
	if !hexColorRegex.MatchString(color) {
		return &ValidationError{
			Field:   "color",
			Message: "Cor deve estar no formato hexadecimal #RRGGBB (ex: #FF5733)",
		}
	}
	return nil
}

// ValidMonth lista de meses válidos em português
var ValidMonths = map[string]bool{
	"Janeiro":   true,
	"Fevereiro": true,
	"Março":     true,
	"Abril":     true,
	"Maio":      true,
	"Junho":     true,
	"Julho":     true,
	"Agosto":    true,
	"Setembro":  true,
	"Outubro":   true,
	"Novembro":  true,
	"Dezembro":  true,
	"Acumulado": true, // Caso especial para férias/13º
}

// ValidateMonth valida se o mês está em português
func ValidateMonth(month string) *ValidationError {
	if month == "" {
		return &ValidationError{
			Field:   "month",
			Message: "Mês não pode ser vazio",
		}
	}

	if !ValidMonths[month] {
		return &ValidationError{
			Field:   "month",
			Message: "Mês inválido. Use mês em português (ex: Janeiro, Fevereiro, etc.)",
		}
	}
	return nil
}

// ValidateStatus valida se o status é "pago" ou "aberto"
func ValidateStatus(status string) *ValidationError {
	statusUpper := strings.ToUpper(status)
	if statusUpper != "PAGO" && statusUpper != "ABERTO" {
		return &ValidationError{
			Field:   "status",
			Message: "Status deve ser 'PAGO' ou 'ABERTO'",
		}
	}
	return nil
}

// ValidateDayOfMonth valida se o dia do mês está entre 1 e 31
func ValidateDayOfMonth(day int) *ValidationError {
	if day < 1 || day > 31 {
		return &ValidationError{
			Field:   "dayOfMonth",
			Message: "Dia do mês deve estar entre 1 e 31",
		}
	}
	return nil
}

// CollectErrors coleta múltiplos erros de validação e retorna apenas os não nulos
func CollectErrors(errors ...*ValidationError) []ValidationError {
	var result []ValidationError
	for _, err := range errors {
		if err != nil {
			result = append(result, *err)
		}
	}
	return result
}

// HasErrors verifica se há erros na lista
func HasErrors(errors []ValidationError) bool {
	return len(errors) > 0
}

// ValidateCategoryType valida o tipo de categoria
func ValidateCategoryType(catType string) *ValidationError {
	if catType == "" {
		return nil
	}
	catTypeUpper := strings.ToUpper(catType)
	if catTypeUpper != "EXPENSE" && catTypeUpper != "INCOME" {
		return &ValidationError{
			Field:   "type",
			Message: "Tipo deve ser 'EXPENSE' ou 'INCOME'",
		}
	}
	return nil
}

// ValidateQuoteStatus valida status do orcamento
func ValidateQuoteStatus(status string) *ValidationError {
	validStatuses := []string{"DRAFT", "SENT", "APPROVED", "REJECTED", "EXECUTED"}
	for _, s := range validStatuses {
		if status == s {
			return nil
		}
	}
	return &ValidationError{
		Field:   "status",
		Message: "Status invalido",
	}
}

// ValidateYear valida se o ano esta dentro de um intervalo valido
func ValidateYear(year int) *ValidationError {
	if year < 2000 || year > 2100 {
		return &ValidationError{
			Field:   "year",
			Message: "Ano deve estar entre 2000 e 2100",
		}
	}
	return nil
}

// ValidateNonNegative valida se um numero e maior ou igual a zero
func ValidateNonNegative(value float64, fieldName string) *ValidationError {
	if value < 0 {
		return &ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("%s deve ser maior ou igual a zero", fieldName),
		}
	}
	return nil
}

// ValidatePercentage valida se um valor de porcentagem esta entre 0 e 100
func ValidatePercentage(value float64, fieldName string) *ValidationError {
	if value < 0 || value > 100 {
		return &ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("%s deve estar entre 0 e 100", fieldName),
		}
	}
	return nil
}

// GetMonthName returns the Portuguese month name for a given month number (1-12)
func GetMonthName(month int) string {
	months := map[int]string{
		1:  "Janeiro",
		2:  "Fevereiro",
		3:  "Marco",
		4:  "Abril",
		5:  "Maio",
		6:  "Junho",
		7:  "Julho",
		8:  "Agosto",
		9:  "Setembro",
		10: "Outubro",
		11: "Novembro",
		12: "Dezembro",
	}
	if name, ok := months[month]; ok {
		return name
	}
	return ""
}

// =============================================================================
// Financial validation constants
// =============================================================================

const (
	// MaxAmount is the maximum allowed monetary value (prevents overflow issues)
	// Using 999,999,999.99 as a sensible limit for financial applications
	MaxAmount = 999999999.99

	// MinAmount is the minimum positive amount allowed
	MinAmount = 0.01

	// MaxDecimalPlaces for monetary values
	MaxDecimalPlaces = 2
)

// =============================================================================
// COMPREHENSIVE FINANCIAL VALIDATION FUNCTIONS
// =============================================================================

// hasExcessiveDecimalPlaces checks if a float has more than the allowed decimal places
// This helps prevent precision issues with monetary values
func hasExcessiveDecimalPlaces(value float64, maxDecimals int) bool {
	// Multiply by 10^maxDecimals and check if there is a fractional part
	multiplier := math.Pow(10, float64(maxDecimals))
	scaled := value * multiplier
	return math.Abs(scaled-math.Round(scaled)) > 1e-9
}

// isValidFiniteNumber checks if a float64 is a valid finite number (not NaN or Inf)
func isValidFiniteNumber(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

// ValidateAmount validates monetary values with comprehensive checks:
// - Must be a valid finite number (not NaN or Infinity)
// - Must be positive (greater than zero)
// - Must not exceed the maximum allowed amount (prevents overflow)
// - Must have at most 2 decimal places (standard for currency)
func ValidateAmount(value float64, fieldName string) *ValidationError {
	// Check for NaN or Infinity (potential attack vector or calculation error)
	if !isValidFiniteNumber(value) {
		return &ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("%s contem um valor numerico invalido", fieldName),
		}
	}

	// Check if positive
	if value <= 0 {
		return &ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("%s deve ser maior que zero", fieldName),
		}
	}

	// Check maximum limit to prevent overflow
	if value > MaxAmount {
		return &ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("%s excede o valor maximo permitido de R$ 999.999.999,99", fieldName),
		}
	}

	// Check decimal places (monetary values should have at most 2 decimal places)
	if hasExcessiveDecimalPlaces(value, MaxDecimalPlaces) {
		return &ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("%s deve ter no maximo 2 casas decimais", fieldName),
		}
	}

	return nil
}

// ValidateAmountAllowZero validates monetary values that can be zero (e.g., budgets):
// - Must be a valid finite number (not NaN or Infinity)
// - Must be non-negative (zero or positive)
// - Must not exceed the maximum allowed amount
// - Must have at most 2 decimal places
func ValidateAmountAllowZero(value float64, fieldName string) *ValidationError {
	// Check for NaN or Infinity
	if !isValidFiniteNumber(value) {
		return &ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("%s contem um valor numerico invalido", fieldName),
		}
	}

	// Check if non-negative
	if value < 0 {
		return &ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("%s deve ser maior ou igual a zero", fieldName),
		}
	}

	// Check maximum limit
	if value > MaxAmount {
		return &ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("%s excede o valor maximo permitido de R$ 999.999.999,99", fieldName),
		}
	}

	// Check decimal places
	if hasExcessiveDecimalPlaces(value, MaxDecimalPlaces) {
		return &ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("%s deve ter no maximo 2 casas decimais", fieldName),
		}
	}

	return nil
}

// ValidateDiscountValue validates discount values based on the discount type:
// - For "PERCENT" type: must be between 0 and 100
// - For "VALUE" type: must be non-negative and not exceed max amount
// - Must be a valid finite number
// - Must have at most 2 decimal places
func ValidateDiscountValue(value float64, discountType string, fieldName string) *ValidationError {
	// Check for NaN or Infinity
	if !isValidFiniteNumber(value) {
		return &ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("%s contem um valor numerico invalido", fieldName),
		}
	}

	// Check if non-negative
	if value < 0 {
		return &ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("%s deve ser maior ou igual a zero", fieldName),
		}
	}

	// Validate based on discount type
	discountTypeUpper := strings.ToUpper(discountType)
	if discountTypeUpper == "PERCENT" {
		// Percentage discount: 0-100
		if value > 100 {
			return &ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("%s percentual deve estar entre 0 e 100", fieldName),
			}
		}
	} else {
		// Value discount: check max amount
		if value > MaxAmount {
			return &ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("%s excede o valor maximo permitido de R$ 999.999.999,99", fieldName),
			}
		}
	}

	// Check decimal places
	if hasExcessiveDecimalPlaces(value, MaxDecimalPlaces) {
		return &ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("%s deve ter no maximo 2 casas decimais", fieldName),
		}
	}

	return nil
}

// ValidateQuantity validates quantity values (used in quote items):
// - Must be a valid finite number
// - Must be positive
// - Must not exceed a reasonable maximum (prevents overflow in calculations)
// - Can have up to 4 decimal places (for fractional quantities like 0.5 hours)
func ValidateQuantity(value float64, fieldName string) *ValidationError {
	const maxQuantity = 999999.9999
	const maxQuantityDecimals = 4

	// Check for NaN or Infinity
	if !isValidFiniteNumber(value) {
		return &ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("%s contem um valor numerico invalido", fieldName),
		}
	}

	// Check if positive
	if value <= 0 {
		return &ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("%s deve ser maior que zero", fieldName),
		}
	}

	// Check maximum
	if value > maxQuantity {
		return &ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("%s excede o valor maximo permitido", fieldName),
		}
	}

	// Check decimal places (allow up to 4 for quantities)
	if hasExcessiveDecimalPlaces(value, maxQuantityDecimals) {
		return &ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("%s deve ter no maximo %d casas decimais", fieldName, maxQuantityDecimals),
		}
	}

	return nil
}

// ValidateUnitPrice validates unit price values:
// - Must be a valid finite number
// - Must be positive
// - Must not exceed max amount
// - Must have at most 2 decimal places
func ValidateUnitPrice(value float64, fieldName string) *ValidationError {
	return ValidateAmount(value, fieldName)
}

// ValidateBudget validates budget values for categories:
// - Can be zero (no budget set)
// - Must be non-negative
// - Must not exceed max amount
// - Must have at most 2 decimal places
func ValidateBudget(value float64, fieldName string) *ValidationError {
	return ValidateAmountAllowZero(value, fieldName)
}

// RoundToTwoDecimals rounds a float64 to 2 decimal places
// This is useful for normalizing monetary values before storage
func RoundToTwoDecimals(value float64) float64 {
	return math.Round(value*100) / 100
}
