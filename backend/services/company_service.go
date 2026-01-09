package services

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"m2m-backend/helpers"
	"m2m-backend/models"
	"m2m-backend/repositories"
)

var (
	// ErrCompanyNotFound is returned when a company is not found
	ErrCompanyNotFound = errors.New("company not found")

	// ErrUnauthorized is returned when user is not authorized
	ErrUnauthorized = errors.New("unauthorized access")

	// ErrInvalidInput is returned when input validation fails
	ErrInvalidInput = errors.New("invalid input")
)

// CompanyService contains business logic for company operations
type CompanyService struct {
	repo *repositories.CompanyRepository
}

// NewCompanyService creates a new instance of CompanyService
func NewCompanyService() *CompanyService {
	return &CompanyService{
		repo: repositories.NewCompanyRepository(),
	}
}

// GetCompaniesByUserID retrieves all companies owned by a user
func (s *CompanyService) GetCompaniesByUserID(userID primitive.ObjectID) ([]models.Company, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	companies, err := s.repo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return companies, nil
}

// CreateCompany creates a new company with validation
func (s *CompanyService) CreateCompany(userID primitive.ObjectID, req models.CreateCompanyRequest) (*models.Company, []helpers.ValidationError, error) {
	// Validate input
	validationErrors := helpers.CollectErrors(
		helpers.ValidateRequired(req.Name, "name"),
		helpers.ValidateMaxLength(req.Name, "name", 100),
		helpers.ValidateNoScriptTags(req.Name, "name"),
		helpers.ValidateMongoInjection(req.Name, "name"),
		helpers.ValidateSQLInjection(req.Name, "name"),
	)

	if helpers.HasErrors(validationErrors) {
		return nil, validationErrors, ErrInvalidInput
	}

	// Sanitize user input
	sanitizedName := helpers.SanitizeString(req.Name)

	// Create company model
	company := &models.Company{
		ID:        primitive.NewObjectID(),
		UserID:    userID,
		Name:      sanitizedName,
		CreatedAt: time.Now(),
	}

	// Save to database
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := s.repo.Create(ctx, company)
	if err != nil {
		return nil, nil, err
	}

	return company, nil, nil
}

// UpdateCompany updates an existing company with validation
func (s *CompanyService) UpdateCompany(companyID, userID primitive.ObjectID, req models.CreateCompanyRequest) (*models.Company, []helpers.ValidationError, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Validate ownership
	company, err := s.repo.ValidateOwnership(ctx, companyID, userID)
	if err != nil {
		if err == repositories.ErrNotFound {
			return nil, nil, ErrCompanyNotFound
		}
		if err == repositories.ErrUnauthorized {
			return nil, nil, ErrUnauthorized
		}
		return nil, nil, err
	}
	if company == nil {
		return nil, nil, ErrCompanyNotFound
	}

	// Validate input
	validationErrors := helpers.CollectErrors(
		helpers.ValidateRequired(req.Name, "name"),
		helpers.ValidateMaxLength(req.Name, "name", 100),
		helpers.ValidateNoScriptTags(req.Name, "name"),
		helpers.ValidateMongoInjection(req.Name, "name"),
		helpers.ValidateSQLInjection(req.Name, "name"),
	)

	if helpers.HasErrors(validationErrors) {
		return nil, validationErrors, ErrInvalidInput
	}

	// Sanitize user input
	sanitizedName := helpers.SanitizeString(req.Name)

	// Update in database
	err = s.repo.Update(ctx, companyID, sanitizedName)
	if err != nil {
		return nil, nil, err
	}

	// Fetch updated company
	updatedCompany, err := s.repo.FindByID(ctx, companyID)
	if err != nil {
		return nil, nil, err
	}

	return updatedCompany, nil, nil
}

// DeleteCompany deletes a company and all related data (cascade delete)
func (s *CompanyService) DeleteCompany(companyID, userID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Validate ownership
	company, err := s.repo.ValidateOwnership(ctx, companyID, userID)
	if err != nil {
		if err == repositories.ErrNotFound {
			return ErrCompanyNotFound
		}
		if err == repositories.ErrUnauthorized {
			return ErrUnauthorized
		}
		return err
	}
	if company == nil {
		return ErrCompanyNotFound
	}

	// Delete related data in cascade
	// Delete transactions
	err = s.repo.DeleteRelatedTransactions(ctx, companyID)
	if err != nil {
		return err
	}

	// Delete categories
	err = s.repo.DeleteRelatedCategories(ctx, companyID)
	if err != nil {
		return err
	}

	// Delete recurring transactions
	err = s.repo.DeleteRelatedRecurring(ctx, companyID)
	if err != nil {
		return err
	}

	// Finally, delete the company itself
	err = s.repo.Delete(ctx, companyID)
	if err != nil {
		return err
	}

	return nil
}

// ValidateCompanyOwnership validates that a user owns a specific company
func (s *CompanyService) ValidateCompanyOwnership(companyID, userID primitive.ObjectID) (*models.Company, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	company, err := s.repo.ValidateOwnership(ctx, companyID, userID)
	if err != nil {
		if err == repositories.ErrNotFound {
			return nil, ErrCompanyNotFound
		}
		if err == repositories.ErrUnauthorized {
			return nil, ErrUnauthorized
		}
		return nil, err
	}
	if company == nil {
		return nil, ErrCompanyNotFound
	}

	return company, nil
}
