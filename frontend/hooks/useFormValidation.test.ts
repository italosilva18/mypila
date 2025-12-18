import { describe, it, expect, beforeEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useFormValidation } from './useFormValidation';
import { ValidationResult } from '../utils/validation';

describe('useFormValidation Hook', () => {
  describe('Initial State', () => {
    it('should initialize with empty errors', () => {
      const { result } = renderHook(() => useFormValidation());

      expect(result.current.errors).toEqual({});
      expect(result.current.hasErrors()).toBe(false);
    });
  });

  describe('validateField', () => {
    it('should validate a single field and return true for valid input', () => {
      const { result } = renderHook(() => useFormValidation());

      let isValid: boolean;
      act(() => {
        isValid = result.current.validateField('email', () => ({
          isValid: true
        }));
      });

      expect(isValid!).toBe(true);
      expect(result.current.errors.email).toEqual({ isValid: true });
      expect(result.current.hasError('email')).toBe(false);
    });

    it('should validate a single field and return false for invalid input', () => {
      const { result } = renderHook(() => useFormValidation());

      let isValid: boolean;
      act(() => {
        isValid = result.current.validateField('password', () => ({
          isValid: false,
          error: 'Senha é obrigatória'
        }));
      });

      expect(isValid!).toBe(false);
      expect(result.current.errors.password).toEqual({
        isValid: false,
        error: 'Senha é obrigatória'
      });
      expect(result.current.hasError('password')).toBe(true);
    });

    it('should update existing field validation', () => {
      const { result } = renderHook(() => useFormValidation());

      act(() => {
        result.current.validateField('username', () => ({
          isValid: false,
          error: 'Nome de usuário inválido'
        }));
      });

      expect(result.current.hasError('username')).toBe(true);

      act(() => {
        result.current.validateField('username', () => ({
          isValid: true
        }));
      });

      expect(result.current.hasError('username')).toBe(false);
    });

    it('should not affect other field errors when validating one field', () => {
      const { result } = renderHook(() => useFormValidation());

      act(() => {
        result.current.validateField('email', () => ({
          isValid: false,
          error: 'Email inválido'
        }));
        result.current.validateField('password', () => ({
          isValid: false,
          error: 'Senha inválida'
        }));
      });

      expect(result.current.errors).toEqual({
        email: { isValid: false, error: 'Email inválido' },
        password: { isValid: false, error: 'Senha inválida' }
      });
    });
  });

  describe('validateFields', () => {
    it('should validate multiple fields and return true when all are valid', () => {
      const { result } = renderHook(() => useFormValidation());

      let allValid: boolean;
      act(() => {
        allValid = result.current.validateFields({
          name: () => ({ isValid: true }),
          email: () => ({ isValid: true }),
          password: () => ({ isValid: true })
        });
      });

      expect(allValid!).toBe(true);
      expect(result.current.hasErrors()).toBe(false);
    });

    it('should validate multiple fields and return false when any is invalid', () => {
      const { result } = renderHook(() => useFormValidation());

      let allValid: boolean;
      act(() => {
        allValid = result.current.validateFields({
          name: () => ({ isValid: true }),
          email: () => ({ isValid: false, error: 'Email inválido' }),
          password: () => ({ isValid: true })
        });
      });

      expect(allValid!).toBe(false);
      expect(result.current.hasErrors()).toBe(true);
      expect(result.current.hasError('email')).toBe(true);
      expect(result.current.hasError('name')).toBe(false);
    });

    it('should set all field errors correctly', () => {
      const { result } = renderHook(() => useFormValidation());

      act(() => {
        result.current.validateFields({
          username: () => ({ isValid: false, error: 'Usuário muito curto' }),
          email: () => ({ isValid: false, error: 'Email inválido' }),
          password: () => ({ isValid: false, error: 'Senha muito fraca' })
        });
      });

      expect(result.current.errors).toEqual({
        username: { isValid: false, error: 'Usuário muito curto' },
        email: { isValid: false, error: 'Email inválido' },
        password: { isValid: false, error: 'Senha muito fraca' }
      });
    });

    it('should replace previous errors with new validation results', () => {
      const { result } = renderHook(() => useFormValidation());

      act(() => {
        result.current.validateField('oldField', () => ({
          isValid: false,
          error: 'Erro antigo'
        }));
      });

      act(() => {
        result.current.validateFields({
          newField: () => ({ isValid: false, error: 'Novo erro' })
        });
      });

      expect(result.current.errors).toEqual({
        newField: { isValid: false, error: 'Novo erro' }
      });
      expect(result.current.errors.oldField).toBeUndefined();
    });
  });

  describe('clearError', () => {
    it('should clear error for specific field', () => {
      const { result } = renderHook(() => useFormValidation());

      act(() => {
        result.current.validateFields({
          email: () => ({ isValid: false, error: 'Email inválido' }),
          password: () => ({ isValid: false, error: 'Senha inválida' })
        });
      });

      expect(result.current.hasError('email')).toBe(true);
      expect(result.current.hasError('password')).toBe(true);

      act(() => {
        result.current.clearError('email');
      });

      expect(result.current.hasError('email')).toBe(false);
      expect(result.current.hasError('password')).toBe(true);
    });

    it('should do nothing if field does not exist', () => {
      const { result } = renderHook(() => useFormValidation());

      act(() => {
        result.current.validateField('email', () => ({
          isValid: false,
          error: 'Erro'
        }));
      });

      const errorsBefore = { ...result.current.errors };

      act(() => {
        result.current.clearError('nonexistent');
      });

      expect(result.current.errors).toEqual(errorsBefore);
    });
  });

  describe('clearAllErrors', () => {
    it('should clear all errors', () => {
      const { result } = renderHook(() => useFormValidation());

      act(() => {
        result.current.validateFields({
          name: () => ({ isValid: false, error: 'Nome inválido' }),
          email: () => ({ isValid: false, error: 'Email inválido' }),
          password: () => ({ isValid: false, error: 'Senha inválida' })
        });
      });

      expect(result.current.hasErrors()).toBe(true);

      act(() => {
        result.current.clearAllErrors();
      });

      expect(result.current.errors).toEqual({});
      expect(result.current.hasErrors()).toBe(false);
    });

    it('should work on empty errors object', () => {
      const { result } = renderHook(() => useFormValidation());

      act(() => {
        result.current.clearAllErrors();
      });

      expect(result.current.errors).toEqual({});
    });
  });

  describe('getError', () => {
    it('should return error message for field with error', () => {
      const { result } = renderHook(() => useFormValidation());

      act(() => {
        result.current.validateField('email', () => ({
          isValid: false,
          error: 'Email é obrigatório'
        }));
      });

      expect(result.current.getError('email')).toBe('Email é obrigatório');
    });

    it('should return undefined for field without error', () => {
      const { result } = renderHook(() => useFormValidation());

      act(() => {
        result.current.validateField('email', () => ({
          isValid: true
        }));
      });

      expect(result.current.getError('email')).toBeUndefined();
    });

    it('should return undefined for nonexistent field', () => {
      const { result } = renderHook(() => useFormValidation());

      expect(result.current.getError('nonexistent')).toBeUndefined();
    });
  });

  describe('hasErrors', () => {
    it('should return false when no errors exist', () => {
      const { result } = renderHook(() => useFormValidation());

      expect(result.current.hasErrors()).toBe(false);
    });

    it('should return false when all fields are valid', () => {
      const { result } = renderHook(() => useFormValidation());

      act(() => {
        result.current.validateFields({
          name: () => ({ isValid: true }),
          email: () => ({ isValid: true })
        });
      });

      expect(result.current.hasErrors()).toBe(false);
    });

    it('should return true when at least one field is invalid', () => {
      const { result } = renderHook(() => useFormValidation());

      act(() => {
        result.current.validateFields({
          name: () => ({ isValid: true }),
          email: () => ({ isValid: false, error: 'Erro' })
        });
      });

      expect(result.current.hasErrors()).toBe(true);
    });
  });

  describe('hasError', () => {
    it('should return true for field with error', () => {
      const { result } = renderHook(() => useFormValidation());

      act(() => {
        result.current.validateField('password', () => ({
          isValid: false,
          error: 'Senha muito curta'
        }));
      });

      expect(result.current.hasError('password')).toBe(true);
    });

    it('should return false for field without error', () => {
      const { result } = renderHook(() => useFormValidation());

      act(() => {
        result.current.validateField('email', () => ({
          isValid: true
        }));
      });

      expect(result.current.hasError('email')).toBe(false);
    });

    it('should return false for nonexistent field', () => {
      const { result } = renderHook(() => useFormValidation());

      expect(result.current.hasError('nonexistent')).toBe(false);
    });
  });

  describe('Complex Scenarios', () => {
    it('should handle complete form validation workflow', () => {
      const { result } = renderHook(() => useFormValidation());

      // Initial validation - all invalid
      act(() => {
        result.current.validateFields({
          name: () => ({ isValid: false, error: 'Nome é obrigatório' }),
          email: () => ({ isValid: false, error: 'Email é obrigatório' }),
          password: () => ({ isValid: false, error: 'Senha é obrigatória' })
        });
      });

      expect(result.current.hasErrors()).toBe(true);

      // User fixes name
      act(() => {
        result.current.validateField('name', () => ({ isValid: true }));
      });

      expect(result.current.hasError('name')).toBe(false);
      expect(result.current.hasErrors()).toBe(true);

      // User fixes email
      act(() => {
        result.current.validateField('email', () => ({ isValid: true }));
      });

      expect(result.current.hasErrors()).toBe(true);

      // User fixes password
      act(() => {
        result.current.validateField('password', () => ({ isValid: true }));
      });

      expect(result.current.hasErrors()).toBe(false);
    });

    it('should handle clearing errors during editing', () => {
      const { result } = renderHook(() => useFormValidation());

      // Validate with errors
      act(() => {
        result.current.validateFields({
          email: () => ({ isValid: false, error: 'Email inválido' })
        });
      });

      // User starts editing - clear error
      act(() => {
        result.current.clearError('email');
      });

      expect(result.current.getError('email')).toBeUndefined();

      // Revalidate on blur
      act(() => {
        result.current.validateField('email', () => ({
          isValid: true
        }));
      });

      expect(result.current.hasError('email')).toBe(false);
    });

    it('should handle form reset', () => {
      const { result } = renderHook(() => useFormValidation());

      // Add errors
      act(() => {
        result.current.validateFields({
          field1: () => ({ isValid: false, error: 'Erro 1' }),
          field2: () => ({ isValid: false, error: 'Erro 2' }),
          field3: () => ({ isValid: false, error: 'Erro 3' })
        });
      });

      expect(Object.keys(result.current.errors).length).toBe(3);

      // Reset form
      act(() => {
        result.current.clearAllErrors();
      });

      expect(result.current.errors).toEqual({});
      expect(result.current.hasErrors()).toBe(false);
    });
  });
});
