package helpers

import (
	"log"

	"github.com/gofiber/fiber/v2"
)

// Error codes for structured error responses
const (
	// General errors
	ErrCodeInternalError    = "INTERNAL_ERROR"
	ErrCodeBadRequest       = "BAD_REQUEST"
	ErrCodeUnauthorized     = "UNAUTHORIZED"
	ErrCodeForbidden        = "FORBIDDEN"
	ErrCodeNotFound         = "NOT_FOUND"
	ErrCodeConflict         = "CONFLICT"
	ErrCodeValidationFailed = "VALIDATION_FAILED"
	ErrCodeRateLimited      = "RATE_LIMITED"

	// Transaction errors
	ErrCodeTransactionNotFound   = "TRANSACTION_NOT_FOUND"
	ErrCodeTransactionCreateFail = "TRANSACTION_CREATE_FAILED"
	ErrCodeTransactionUpdateFail = "TRANSACTION_UPDATE_FAILED"
	ErrCodeTransactionDeleteFail = "TRANSACTION_DELETE_FAILED"
	ErrCodeTransactionFetchFail  = "TRANSACTION_FETCH_FAILED"

	// Category errors
	ErrCodeCategoryNotFound   = "CATEGORY_NOT_FOUND"
	ErrCodeCategoryCreateFail = "CATEGORY_CREATE_FAILED"
	ErrCodeCategoryUpdateFail = "CATEGORY_UPDATE_FAILED"
	ErrCodeCategoryDeleteFail = "CATEGORY_DELETE_FAILED"
	ErrCodeCategoryFetchFail  = "CATEGORY_FETCH_FAILED"

	// Company errors
	ErrCodeCompanyNotFound   = "COMPANY_NOT_FOUND"
	ErrCodeCompanyCreateFail = "COMPANY_CREATE_FAILED"
	ErrCodeCompanyUpdateFail = "COMPANY_UPDATE_FAILED"
	ErrCodeCompanyDeleteFail = "COMPANY_DELETE_FAILED"
	ErrCodeCompanyFetchFail  = "COMPANY_FETCH_FAILED"

	// Quote errors
	ErrCodeQuoteNotFound      = "QUOTE_NOT_FOUND"
	ErrCodeQuoteCreateFail    = "QUOTE_CREATE_FAILED"
	ErrCodeQuoteUpdateFail    = "QUOTE_UPDATE_FAILED"
	ErrCodeQuoteDeleteFail    = "QUOTE_DELETE_FAILED"
	ErrCodeQuoteFetchFail     = "QUOTE_FETCH_FAILED"
	ErrCodeQuoteDuplicateFail = "QUOTE_DUPLICATE_FAILED"
	ErrCodeQuoteExecuted      = "QUOTE_ALREADY_EXECUTED"

	// Quote template errors
	ErrCodeQuoteTemplateNotFound   = "QUOTE_TEMPLATE_NOT_FOUND"
	ErrCodeQuoteTemplateCreateFail = "QUOTE_TEMPLATE_CREATE_FAILED"
	ErrCodeQuoteTemplateUpdateFail = "QUOTE_TEMPLATE_UPDATE_FAILED"
	ErrCodeQuoteTemplateDeleteFail = "QUOTE_TEMPLATE_DELETE_FAILED"
	ErrCodeQuoteTemplateFetchFail  = "QUOTE_TEMPLATE_FETCH_FAILED"

	// Recurring errors
	ErrCodeRecurringNotFound    = "RECURRING_NOT_FOUND"
	ErrCodeRecurringCreateFail  = "RECURRING_CREATE_FAILED"
	ErrCodeRecurringDeleteFail  = "RECURRING_DELETE_FAILED"
	ErrCodeRecurringFetchFail   = "RECURRING_FETCH_FAILED"
	ErrCodeRecurringProcessFail = "RECURRING_PROCESS_FAILED"

	// Auth errors
	ErrCodeAuthInvalidCredentials = "INVALID_CREDENTIALS"
	ErrCodeAuthEmailExists        = "EMAIL_ALREADY_EXISTS"
	ErrCodeAuthTokenFailed        = "TOKEN_GENERATION_FAILED"
	ErrCodeAuthUserNotFound       = "USER_NOT_FOUND"
	ErrCodeAuthInvalidToken       = "INVALID_TOKEN"

	// Database errors
	ErrCodeDatabaseError    = "DATABASE_ERROR"
	ErrCodeTransactionError = "TRANSACTION_ERROR"

	// ID format errors
	ErrCodeInvalidID = "INVALID_ID_FORMAT"
)

// ErrorDetails contains optional additional error information
type ErrorDetails map[string]interface{}

// ErrorBody represents the error object in the response
type ErrorBody struct {
	Code    string       `json:"code"`
	Message string       `json:"message"`
	Details ErrorDetails `json:"details,omitempty"`
}

// ErrorResponse represents a structured error response with request ID
type ErrorResponse struct {
	RequestID string    `json:"requestId"`
	Error     ErrorBody `json:"error"`
}

// GetRequestID extracts the request ID from fiber context
// The request ID is set by the requestid middleware
func GetRequestID(c *fiber.Ctx) string {
	if reqID := c.Locals("requestid"); reqID != nil {
		if id, ok := reqID.(string); ok {
			return id
		}
	}
	return "unknown"
}

// logError logs the error with request context for debugging
func logError(c *fiber.Ctx, code, message string, err error) {
	requestID := GetRequestID(c)
	if err != nil {
		log.Printf("[ERROR] requestId=%s code=%s message=%s error=%v", requestID, code, message, err)
	} else {
		log.Printf("[ERROR] requestId=%s code=%s message=%s", requestID, code, message)
	}
}

// sendError is the base function that sends structured error responses
func sendError(c *fiber.Ctx, status int, code, message string, details ErrorDetails) error {
	requestID := GetRequestID(c)

	response := ErrorResponse{
		RequestID: requestID,
		Error: ErrorBody{
			Code:    code,
			Message: message,
			Details: details,
		},
	}

	return c.Status(status).JSON(response)
}

// BadRequest sends a 400 Bad Request error response
func BadRequest(c *fiber.Ctx, code, message string, details ErrorDetails) error {
	logError(c, code, message, nil)
	return sendError(c, fiber.StatusBadRequest, code, message, details)
}

// BadRequestWithError sends a 400 Bad Request error response with error logging
func BadRequestWithError(c *fiber.Ctx, code, message string, err error, details ErrorDetails) error {
	logError(c, code, message, err)
	return sendError(c, fiber.StatusBadRequest, code, message, details)
}

// Unauthorized sends a 401 Unauthorized error response
func Unauthorized(c *fiber.Ctx, code, message string, details ErrorDetails) error {
	logError(c, code, message, nil)
	return sendError(c, fiber.StatusUnauthorized, code, message, details)
}

// Forbidden sends a 403 Forbidden error response
func Forbidden(c *fiber.Ctx, code, message string, details ErrorDetails) error {
	logError(c, code, message, nil)
	return sendError(c, fiber.StatusForbidden, code, message, details)
}

// NotFound sends a 404 Not Found error response
func NotFound(c *fiber.Ctx, code, message string, details ErrorDetails) error {
	logError(c, code, message, nil)
	return sendError(c, fiber.StatusNotFound, code, message, details)
}

// Conflict sends a 409 Conflict error response
func Conflict(c *fiber.Ctx, code, message string, details ErrorDetails) error {
	logError(c, code, message, nil)
	return sendError(c, fiber.StatusConflict, code, message, details)
}

// InternalError sends a 500 Internal Server Error response
func InternalError(c *fiber.Ctx, code, message string, err error, details ErrorDetails) error {
	logError(c, code, message, err)
	return sendError(c, fiber.StatusInternalServerError, code, message, details)
}

// RateLimited sends a 429 Too Many Requests error response
func RateLimited(c *fiber.Ctx, message string, details ErrorDetails) error {
	logError(c, ErrCodeRateLimited, message, nil)
	return sendError(c, fiber.StatusTooManyRequests, ErrCodeRateLimited, message, details)
}

// --- Convenience functions for common error scenarios ---

// InvalidIDFormat sends a standardized invalid ID format error
func InvalidIDFormat(c *fiber.Ctx, field string) error {
	return BadRequest(c, ErrCodeInvalidID, "Formato de ID invalido", ErrorDetails{"field": field})
}

// InvalidRequestBody sends a standardized invalid request body error
func InvalidRequestBody(c *fiber.Ctx) error {
	return BadRequest(c, ErrCodeBadRequest, "Corpo da requisicao invalido", nil)
}

// MissingRequiredParam sends a standardized missing required parameter error
func MissingRequiredParam(c *fiber.Ctx, param string) error {
	return BadRequest(c, ErrCodeBadRequest, param+" e obrigatorio", ErrorDetails{"param": param})
}

// DatabaseError sends a standardized database error (without exposing internal details)
func DatabaseError(c *fiber.Ctx, operation string, err error) error {
	return InternalError(c, ErrCodeDatabaseError, "Erro no banco de dados", err, ErrorDetails{"operation": operation})
}

// TransactionError sends a standardized transaction error
func TransactionError(c *fiber.Ctx, operation string, err error) error {
	return InternalError(c, ErrCodeTransactionError, "Falha na transacao atomica", err, ErrorDetails{"operation": operation})
}

// --- Transaction error helpers ---

// TransactionNotFound sends a transaction not found error
func TransactionNotFound(c *fiber.Ctx) error {
	return NotFound(c, ErrCodeTransactionNotFound, "Transacao nao encontrada", nil)
}

// TransactionFetchFailed sends a transaction fetch failed error
func TransactionFetchFailed(c *fiber.Ctx, err error) error {
	return InternalError(c, ErrCodeTransactionFetchFail, "Falha ao buscar transacoes", err, nil)
}

// TransactionCreateFailed sends a transaction create failed error
func TransactionCreateFailed(c *fiber.Ctx, err error) error {
	return InternalError(c, ErrCodeTransactionCreateFail, "Falha ao criar transacao", err, nil)
}

// TransactionUpdateFailed sends a transaction update failed error
func TransactionUpdateFailed(c *fiber.Ctx, err error) error {
	return InternalError(c, ErrCodeTransactionUpdateFail, "Falha ao atualizar transacao", err, nil)
}

// TransactionDeleteFailed sends a transaction delete failed error
func TransactionDeleteFailed(c *fiber.Ctx, err error) error {
	return InternalError(c, ErrCodeTransactionDeleteFail, "Falha ao excluir transacao", err, nil)
}

// --- Category error helpers ---

// CategoryNotFound sends a category not found error
func CategoryNotFound(c *fiber.Ctx) error {
	return NotFound(c, ErrCodeCategoryNotFound, "Categoria nao encontrada", nil)
}

// CategoryFetchFailed sends a category fetch failed error
func CategoryFetchFailed(c *fiber.Ctx, err error) error {
	return InternalError(c, ErrCodeCategoryFetchFail, "Falha ao buscar categorias", err, nil)
}

// CategoryCreateFailed sends a category create failed error
func CategoryCreateFailed(c *fiber.Ctx, err error) error {
	return InternalError(c, ErrCodeCategoryCreateFail, "Falha ao criar categoria", err, nil)
}

// CategoryUpdateFailed sends a category update failed error
func CategoryUpdateFailed(c *fiber.Ctx, err error) error {
	return InternalError(c, ErrCodeCategoryUpdateFail, "Falha ao atualizar categoria", err, nil)
}

// CategoryDeleteFailed sends a category delete failed error
func CategoryDeleteFailed(c *fiber.Ctx, err error) error {
	return InternalError(c, ErrCodeCategoryDeleteFail, "Falha ao excluir categoria", err, nil)
}

// --- Company error helpers ---

// CompanyNotFound sends a company not found error
func CompanyNotFound(c *fiber.Ctx) error {
	return NotFound(c, ErrCodeCompanyNotFound, "Empresa nao encontrada", nil)
}

// CompanyFetchFailed sends a company fetch failed error
func CompanyFetchFailed(c *fiber.Ctx, err error) error {
	return InternalError(c, ErrCodeCompanyFetchFail, "Falha ao buscar empresas", err, nil)
}

// CompanyCreateFailed sends a company create failed error
func CompanyCreateFailed(c *fiber.Ctx, err error) error {
	return InternalError(c, ErrCodeCompanyCreateFail, "Falha ao criar empresa", err, nil)
}

// CompanyUpdateFailed sends a company update failed error
func CompanyUpdateFailed(c *fiber.Ctx, err error) error {
	return InternalError(c, ErrCodeCompanyUpdateFail, "Falha ao atualizar empresa", err, nil)
}

// CompanyDeleteFailed sends a company delete failed error
func CompanyDeleteFailed(c *fiber.Ctx, err error) error {
	return InternalError(c, ErrCodeCompanyDeleteFail, "Falha ao excluir empresa", err, nil)
}

// --- Quote error helpers ---

// QuoteNotFound sends a quote not found error
func QuoteNotFound(c *fiber.Ctx) error {
	return NotFound(c, ErrCodeQuoteNotFound, "Orcamento nao encontrado", nil)
}

// QuoteFetchFailed sends a quote fetch failed error
func QuoteFetchFailed(c *fiber.Ctx, err error) error {
	return InternalError(c, ErrCodeQuoteFetchFail, "Falha ao buscar orcamentos", err, nil)
}

// QuoteCreateFailed sends a quote create failed error
func QuoteCreateFailed(c *fiber.Ctx, err error) error {
	return InternalError(c, ErrCodeQuoteCreateFail, "Falha ao criar orcamento", err, nil)
}

// QuoteUpdateFailed sends a quote update failed error
func QuoteUpdateFailed(c *fiber.Ctx, err error) error {
	return InternalError(c, ErrCodeQuoteUpdateFail, "Falha ao atualizar orcamento", err, nil)
}

// QuoteDeleteFailed sends a quote delete failed error
func QuoteDeleteFailed(c *fiber.Ctx, err error) error {
	return InternalError(c, ErrCodeQuoteDeleteFail, "Falha ao excluir orcamento", err, nil)
}

// QuoteDuplicateFailed sends a quote duplicate failed error
func QuoteDuplicateFailed(c *fiber.Ctx, err error) error {
	return InternalError(c, ErrCodeQuoteDuplicateFail, "Falha ao duplicar orcamento", err, nil)
}

// QuoteAlreadyExecuted sends a quote already executed error
func QuoteAlreadyExecuted(c *fiber.Ctx) error {
	return BadRequest(c, ErrCodeQuoteExecuted, "Nao e possivel editar um orcamento executado", nil)
}

// QuoteNumberGenerationFailed sends a quote number generation failed error
func QuoteNumberGenerationFailed(c *fiber.Ctx, err error) error {
	return InternalError(c, ErrCodeQuoteCreateFail, "Falha ao gerar numero do orcamento", err, nil)
}

// --- Quote template error helpers ---

// QuoteTemplateNotFound sends a quote template not found error
func QuoteTemplateNotFound(c *fiber.Ctx) error {
	return NotFound(c, ErrCodeQuoteTemplateNotFound, "Template de orcamento nao encontrado", nil)
}

// QuoteTemplateFetchFailed sends a quote template fetch failed error
func QuoteTemplateFetchFailed(c *fiber.Ctx, err error) error {
	return InternalError(c, ErrCodeQuoteTemplateFetchFail, "Falha ao buscar templates", err, nil)
}

// QuoteTemplateCreateFailed sends a quote template create failed error
func QuoteTemplateCreateFailed(c *fiber.Ctx, err error) error {
	return InternalError(c, ErrCodeQuoteTemplateCreateFail, "Falha ao criar template", err, nil)
}

// QuoteTemplateUpdateFailed sends a quote template update failed error
func QuoteTemplateUpdateFailed(c *fiber.Ctx, err error) error {
	return InternalError(c, ErrCodeQuoteTemplateUpdateFail, "Falha ao atualizar template", err, nil)
}

// QuoteTemplateDeleteFailed sends a quote template delete failed error
func QuoteTemplateDeleteFailed(c *fiber.Ctx, err error) error {
	return InternalError(c, ErrCodeQuoteTemplateDeleteFail, "Falha ao excluir template", err, nil)
}

// --- Recurring error helpers ---

// RecurringNotFound sends a recurring not found error
func RecurringNotFound(c *fiber.Ctx) error {
	return NotFound(c, ErrCodeRecurringNotFound, "Regra recorrente nao encontrada", nil)
}

// RecurringFetchFailed sends a recurring fetch failed error
func RecurringFetchFailed(c *fiber.Ctx, err error) error {
	return InternalError(c, ErrCodeRecurringFetchFail, "Falha ao buscar regras recorrentes", err, nil)
}

// RecurringCreateFailed sends a recurring create failed error
func RecurringCreateFailed(c *fiber.Ctx, err error) error {
	return InternalError(c, ErrCodeRecurringCreateFail, "Falha ao criar regra recorrente", err, nil)
}

// RecurringDeleteFailed sends a recurring delete failed error
func RecurringDeleteFailed(c *fiber.Ctx, err error) error {
	return InternalError(c, ErrCodeRecurringDeleteFail, "Falha ao excluir regra recorrente", err, nil)
}

// RecurringProcessFailed sends a recurring process failed error
func RecurringProcessFailed(c *fiber.Ctx, err error) error {
	return InternalError(c, ErrCodeRecurringProcessFail, "Falha ao processar regras recorrentes", err, nil)
}

// --- Auth error helpers ---

// AuthInvalidCredentials sends an invalid credentials error
func AuthInvalidCredentials(c *fiber.Ctx) error {
	return Unauthorized(c, ErrCodeAuthInvalidCredentials, "Credenciais invalidas", nil)
}

// AuthEmailExists sends an email already exists error
func AuthEmailExists(c *fiber.Ctx) error {
	return Conflict(c, ErrCodeAuthEmailExists, "Email ja cadastrado", nil)
}

// AuthTokenFailed sends a token generation failed error
func AuthTokenFailed(c *fiber.Ctx, err error) error {
	return InternalError(c, ErrCodeAuthTokenFailed, "Falha ao gerar token", err, nil)
}

// AuthUserNotFound sends a user not found error
func AuthUserNotFound(c *fiber.Ctx) error {
	return NotFound(c, ErrCodeAuthUserNotFound, "Usuario nao encontrado", nil)
}

// AuthInvalidToken sends an invalid token error
func AuthInvalidToken(c *fiber.Ctx) error {
	return Unauthorized(c, ErrCodeAuthInvalidToken, "Token invalido ou expirado", nil)
}

// AuthInvalidUserContext sends an invalid user context error
func AuthInvalidUserContext(c *fiber.Ctx) error {
	return Unauthorized(c, ErrCodeUnauthorized, "Contexto de usuario invalido", nil)
}
