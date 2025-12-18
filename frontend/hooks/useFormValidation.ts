import { useState, useCallback } from 'react';
import { ValidationResult } from '../utils/validation';

export interface FieldValidation {
  [fieldName: string]: ValidationResult;
}

/**
 * Custom hook for managing form validation state
 */
export const useFormValidation = () => {
  const [errors, setErrors] = useState<FieldValidation>({});

  /**
   * Validates a single field
   */
  const validateField = useCallback((
    fieldName: string,
    validationFn: () => ValidationResult
  ): boolean => {
    const result = validationFn();

    setErrors(prev => ({
      ...prev,
      [fieldName]: result
    }));

    return result.isValid;
  }, []);

  /**
   * Validates multiple fields at once
   */
  const validateFields = useCallback((
    validations: { [fieldName: string]: () => ValidationResult }
  ): boolean => {
    const results: FieldValidation = {};
    let allValid = true;

    Object.entries(validations).forEach(([fieldName, validationFn]) => {
      const result = validationFn();
      results[fieldName] = result;
      if (!result.isValid) {
        allValid = false;
      }
    });

    setErrors(results);
    return allValid;
  }, []);

  /**
   * Clears error for a specific field
   */
  const clearError = useCallback((fieldName: string) => {
    setErrors(prev => {
      const newErrors = { ...prev };
      delete newErrors[fieldName];
      return newErrors;
    });
  }, []);

  /**
   * Clears all errors
   */
  const clearAllErrors = useCallback(() => {
    setErrors({});
  }, []);

  /**
   * Gets error message for a field
   */
  const getError = useCallback((fieldName: string): string | undefined => {
    return errors[fieldName]?.error;
  }, [errors]);

  /**
   * Checks if form has any errors
   */
  const hasErrors = useCallback((): boolean => {
    return Object.values(errors).some((result: ValidationResult) => !result.isValid);
  }, [errors]);

  /**
   * Checks if a specific field has error
   */
  const hasError = useCallback((fieldName: string): boolean => {
    return errors[fieldName] && !errors[fieldName].isValid;
  }, [errors]);

  return {
    errors,
    validateField,
    validateFields,
    clearError,
    clearAllErrors,
    getError,
    hasErrors,
    hasError
  };
};
