package helpers

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/gofiber/fiber/v2"
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

	// Regex simples para validação de email
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
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

	// Regex para validar formato #RRGGBB
	hexColorRegex := regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)
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
