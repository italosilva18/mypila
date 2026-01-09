/**
 * Validation utilities for form inputs
 */

export interface ValidationResult {
  isValid: boolean;
  error?: string;
}

/**
 * Validates required field
 */
export const validateRequired = (value: unknown, fieldName: string): ValidationResult => {
  if (value === null || value === undefined || value === '') {
    return {
      isValid: false,
      error: `${fieldName} é obrigatório`
    };
  }

  if (typeof value === 'string' && value.trim() === '') {
    return {
      isValid: false,
      error: `${fieldName} é obrigatório`
    };
  }

  return { isValid: true };
};

/**
 * Validates minimum length for strings
 */
export const validateMinLength = (value: string, min: number, fieldName: string): ValidationResult => {
  if (!value) {
    return { isValid: true };
  }

  if (value.length < min) {
    return {
      isValid: false,
      error: `${fieldName} deve ter no mínimo ${min} caracteres`
    };
  }

  return { isValid: true };
};

/**
 * Validates maximum length for strings
 */
export const validateMaxLength = (value: string, max: number, fieldName: string): ValidationResult => {
  if (!value) {
    return { isValid: true };
  }

  if (value.length > max) {
    return {
      isValid: false,
      error: `${fieldName} deve ter no máximo ${max} caracteres`
    };
  }

  return { isValid: true };
};

/**
 * Validates positive number
 */
export const validatePositiveNumber = (value: number | string, fieldName: string): ValidationResult => {
  const numValue = typeof value === 'string' ? parseFloat(value) : value;

  if (isNaN(numValue)) {
    return {
      isValid: false,
      error: `${fieldName} deve ser um número válido`
    };
  }

  if (numValue <= 0) {
    return {
      isValid: false,
      error: `${fieldName} deve ser maior que zero`
    };
  }

  return { isValid: true };
};

/**
 * Validates number within a range
 */
export const validateRange = (value: number | string, min: number, max: number, fieldName: string): ValidationResult => {
  const numValue = typeof value === 'string' ? parseFloat(value) : value;

  if (isNaN(numValue)) {
    return {
      isValid: false,
      error: `${fieldName} deve ser um número válido`
    };
  }

  if (numValue < min || numValue > max) {
    return {
      isValid: false,
      error: `${fieldName} deve estar entre ${min} e ${max}`
    };
  }

  return { isValid: true };
};

/**
 * Combines multiple validation results
 */
export const combineValidations = (...results: ValidationResult[]): ValidationResult => {
  const firstError = results.find(r => !r.isValid);
  return firstError || { isValid: true };
};
