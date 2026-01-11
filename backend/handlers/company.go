package handlers

import (
	"sync"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"m2m-backend/helpers"
	"m2m-backend/models"
	"m2m-backend/services"
)

var (
	companyService     *services.CompanyService
	companyServiceOnce sync.Once
)

func getCompanyService() *services.CompanyService {
	companyServiceOnce.Do(func() {
		companyService = services.NewCompanyService()
	})
	return companyService
}

// ResetCompanyService resets the singleton for testing purposes
// This allows tests to reinitialize the service with a fresh repository
func ResetCompanyService() {
	companyServiceOnce = sync.Once{}
	companyService = nil
}

// GetCompanies returns all companies owned by the authenticated user
func GetCompanies(c *fiber.Ctx) error {
	// Get authenticated user ID
	userID, err := helpers.GetUserIDFromContext(c)
	if err != nil {
		return err
	}

	// Call service layer
	companies, err := getCompanyService().GetCompaniesByUserID(userID)
	if err != nil {
		return helpers.CompanyFetchFailed(c, err)
	}

	return c.JSON(companies)
}

// CreateCompany creates a new company for the authenticated user
func CreateCompany(c *fiber.Ctx) error {
	// Get authenticated user ID
	userID, err := helpers.GetUserIDFromContext(c)
	if err != nil {
		return err
	}

	// Parse request body
	var req models.CreateCompanyRequest
	if err := c.BodyParser(&req); err != nil {
		return helpers.InvalidRequestBody(c)
	}

	// Call service layer
	company, validationErrors, err := getCompanyService().CreateCompany(userID, req)
	if err != nil {
		if err == services.ErrInvalidInput && validationErrors != nil {
			return helpers.SendValidationErrors(c, validationErrors)
		}
		return helpers.CompanyCreateFailed(c, err)
	}

	return c.Status(201).JSON(company)
}

// UpdateCompany updates an existing company (ownership validated)
func UpdateCompany(c *fiber.Ctx) error {
	// Get authenticated user ID
	userID, err := helpers.GetUserIDFromContext(c)
	if err != nil {
		return err
	}

	// Parse and validate company ID
	id := c.Params("id")
	companyID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return helpers.InvalidIDFormat(c, "id")
	}

	// Parse request body
	var req models.CreateCompanyRequest
	if err := c.BodyParser(&req); err != nil {
		return helpers.InvalidRequestBody(c)
	}

	// Call service layer
	company, validationErrors, err := getCompanyService().UpdateCompany(companyID, userID, req)
	if err != nil {
		if err == services.ErrInvalidInput && validationErrors != nil {
			return helpers.SendValidationErrors(c, validationErrors)
		}
		if err == services.ErrCompanyNotFound {
			return helpers.CompanyNotFound(c)
		}
		if err == services.ErrUnauthorized {
			return helpers.Forbidden(c, helpers.ErrCodeForbidden, "Voce nao tem permissao para acessar esta empresa", helpers.ErrorDetails{
				"reason": "Esta empresa pertence a outro usuario",
			})
		}
		return helpers.CompanyUpdateFailed(c, err)
	}

	return c.JSON(company)
}

// DeleteCompany deletes a company and all related data in cascade (ownership validated)
func DeleteCompany(c *fiber.Ctx) error {
	// Get authenticated user ID
	userID, err := helpers.GetUserIDFromContext(c)
	if err != nil {
		return err
	}

	// Parse and validate company ID
	id := c.Params("id")
	companyID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return helpers.InvalidIDFormat(c, "id")
	}

	// Call service layer
	err = getCompanyService().DeleteCompany(companyID, userID)
	if err != nil {
		if err == services.ErrCompanyNotFound {
			return helpers.CompanyNotFound(c)
		}
		if err == services.ErrUnauthorized {
			return helpers.Forbidden(c, helpers.ErrCodeForbidden, "Voce nao tem permissao para excluir esta empresa", helpers.ErrorDetails{
				"reason": "Esta empresa pertence a outro usuario",
			})
		}
		return helpers.CompanyDeleteFailed(c, err)
	}

	return c.JSON(fiber.Map{"message": "Empresa excluida com sucesso"})
}
