import { describe, it, expect } from 'vitest';
import {
  validateRequired,
  validateMinLength,
  validateMaxLength,
  validatePositiveNumber,
  validateRange,
  combineValidations,
  ValidationResult
} from './validation';

describe('Validation Utils', () => {
  describe('validateRequired', () => {
    it('should return valid for non-empty string', () => {
      const result = validateRequired('test', 'Nome');
      expect(result.isValid).toBe(true);
      expect(result.error).toBeUndefined();
    });

    it('should return invalid for empty string', () => {
      const result = validateRequired('', 'Nome');
      expect(result.isValid).toBe(false);
      expect(result.error).toBe('Nome é obrigatório');
    });

    it('should return invalid for string with only whitespace', () => {
      const result = validateRequired('   ', 'Email');
      expect(result.isValid).toBe(false);
      expect(result.error).toBe('Email é obrigatório');
    });

    it('should return invalid for null value', () => {
      const result = validateRequired(null, 'Senha');
      expect(result.isValid).toBe(false);
      expect(result.error).toBe('Senha é obrigatório');
    });

    it('should return invalid for undefined value', () => {
      const result = validateRequired(undefined, 'Telefone');
      expect(result.isValid).toBe(false);
      expect(result.error).toBe('Telefone é obrigatório');
    });

    it('should return valid for non-empty number', () => {
      const result = validateRequired(123, 'Valor');
      expect(result.isValid).toBe(true);
    });

    it('should return valid for boolean false', () => {
      const result = validateRequired(false, 'Ativo');
      expect(result.isValid).toBe(true);
    });

    it('should return valid for zero', () => {
      const result = validateRequired(0, 'Quantidade');
      expect(result.isValid).toBe(true);
    });
  });

  describe('validateMinLength', () => {
    it('should return valid for string meeting minimum length', () => {
      const result = validateMinLength('password', 8, 'Senha');
      expect(result.isValid).toBe(true);
      expect(result.error).toBeUndefined();
    });

    it('should return valid for string exceeding minimum length', () => {
      const result = validateMinLength('longpassword', 8, 'Senha');
      expect(result.isValid).toBe(true);
    });

    it('should return invalid for string below minimum length', () => {
      const result = validateMinLength('short', 8, 'Senha');
      expect(result.isValid).toBe(false);
      expect(result.error).toBe('Senha deve ter no mínimo 8 caracteres');
    });

    it('should return valid for empty string (optional field)', () => {
      const result = validateMinLength('', 5, 'Descrição');
      expect(result.isValid).toBe(true);
    });

    it('should return valid for null value', () => {
      const result = validateMinLength(null as any, 5, 'Descrição');
      expect(result.isValid).toBe(true);
    });

    it('should return valid for undefined value', () => {
      const result = validateMinLength(undefined as any, 5, 'Descrição');
      expect(result.isValid).toBe(true);
    });

    it('should return valid for string exactly at minimum length', () => {
      const result = validateMinLength('12345', 5, 'Código');
      expect(result.isValid).toBe(true);
    });
  });

  describe('validateMaxLength', () => {
    it('should return valid for string below maximum length', () => {
      const result = validateMaxLength('short', 20, 'Nome');
      expect(result.isValid).toBe(true);
      expect(result.error).toBeUndefined();
    });

    it('should return invalid for string exceeding maximum length', () => {
      const result = validateMaxLength('this is a very long text', 10, 'Título');
      expect(result.isValid).toBe(false);
      expect(result.error).toBe('Título deve ter no máximo 10 caracteres');
    });

    it('should return valid for string exactly at maximum length', () => {
      const result = validateMaxLength('exact', 5, 'Campo');
      expect(result.isValid).toBe(true);
    });

    it('should return valid for empty string', () => {
      const result = validateMaxLength('', 10, 'Descrição');
      expect(result.isValid).toBe(true);
    });

    it('should return valid for null value', () => {
      const result = validateMaxLength(null as any, 10, 'Descrição');
      expect(result.isValid).toBe(true);
    });

    it('should return valid for undefined value', () => {
      const result = validateMaxLength(undefined as any, 10, 'Descrição');
      expect(result.isValid).toBe(true);
    });
  });

  describe('validatePositiveNumber', () => {
    it('should return valid for positive number', () => {
      const result = validatePositiveNumber(100, 'Valor');
      expect(result.isValid).toBe(true);
      expect(result.error).toBeUndefined();
    });

    it('should return valid for positive decimal number', () => {
      const result = validatePositiveNumber(99.99, 'Preço');
      expect(result.isValid).toBe(true);
    });

    it('should return valid for positive string number', () => {
      const result = validatePositiveNumber('150.50', 'Valor');
      expect(result.isValid).toBe(true);
    });

    it('should return invalid for zero', () => {
      const result = validatePositiveNumber(0, 'Quantidade');
      expect(result.isValid).toBe(false);
      expect(result.error).toBe('Quantidade deve ser maior que zero');
    });

    it('should return invalid for negative number', () => {
      const result = validatePositiveNumber(-50, 'Saldo');
      expect(result.isValid).toBe(false);
      expect(result.error).toBe('Saldo deve ser maior que zero');
    });

    it('should return invalid for negative string number', () => {
      const result = validatePositiveNumber('-10.5', 'Valor');
      expect(result.isValid).toBe(false);
      expect(result.error).toBe('Valor deve ser maior que zero');
    });

    it('should return invalid for NaN string', () => {
      const result = validatePositiveNumber('abc', 'Preço');
      expect(result.isValid).toBe(false);
      expect(result.error).toBe('Preço deve ser um número válido');
    });

    it('should return invalid for empty string', () => {
      const result = validatePositiveNumber('', 'Total');
      expect(result.isValid).toBe(false);
      expect(result.error).toBe('Total deve ser um número válido');
    });

    it('should return valid for very small positive number', () => {
      const result = validatePositiveNumber(0.01, 'Taxa');
      expect(result.isValid).toBe(true);
    });

    it('should return valid for large positive number', () => {
      const result = validatePositiveNumber(1000000, 'Limite');
      expect(result.isValid).toBe(true);
    });
  });

  describe('validateRange', () => {
    it('should return valid for number within range', () => {
      const result = validateRange(50, 1, 100, 'Idade');
      expect(result.isValid).toBe(true);
      expect(result.error).toBeUndefined();
    });

    it('should return valid for number at minimum boundary', () => {
      const result = validateRange(1, 1, 100, 'Quantidade');
      expect(result.isValid).toBe(true);
    });

    it('should return valid for number at maximum boundary', () => {
      const result = validateRange(100, 1, 100, 'Porcentagem');
      expect(result.isValid).toBe(true);
    });

    it('should return valid for string number within range', () => {
      const result = validateRange('75.5', 0, 100, 'Taxa');
      expect(result.isValid).toBe(true);
    });

    it('should return invalid for number below minimum', () => {
      const result = validateRange(0, 1, 100, 'Score');
      expect(result.isValid).toBe(false);
      expect(result.error).toBe('Score deve estar entre 1 e 100');
    });

    it('should return invalid for number above maximum', () => {
      const result = validateRange(150, 1, 100, 'Porcentagem');
      expect(result.isValid).toBe(false);
      expect(result.error).toBe('Porcentagem deve estar entre 1 e 100');
    });

    it('should return invalid for NaN string', () => {
      const result = validateRange('invalid', 0, 100, 'Valor');
      expect(result.isValid).toBe(false);
      expect(result.error).toBe('Valor deve ser um número válido');
    });

    it('should handle negative ranges', () => {
      const result = validateRange(-50, -100, 0, 'Temperatura');
      expect(result.isValid).toBe(true);
    });

    it('should handle decimal ranges', () => {
      const result = validateRange(0.5, 0, 1, 'Taxa');
      expect(result.isValid).toBe(true);
    });

    it('should return invalid for string number below minimum', () => {
      const result = validateRange('-10', 0, 100, 'Index');
      expect(result.isValid).toBe(false);
      expect(result.error).toBe('Index deve estar entre 0 e 100');
    });
  });

  describe('combineValidations', () => {
    it('should return valid when all validations pass', () => {
      const results: ValidationResult[] = [
        { isValid: true },
        { isValid: true },
        { isValid: true }
      ];
      const combined = combineValidations(...results);
      expect(combined.isValid).toBe(true);
      expect(combined.error).toBeUndefined();
    });

    it('should return first error when one validation fails', () => {
      const results: ValidationResult[] = [
        { isValid: true },
        { isValid: false, error: 'Erro 1' },
        { isValid: false, error: 'Erro 2' }
      ];
      const combined = combineValidations(...results);
      expect(combined.isValid).toBe(false);
      expect(combined.error).toBe('Erro 1');
    });

    it('should return valid for empty array', () => {
      const combined = combineValidations();
      expect(combined.isValid).toBe(true);
    });

    it('should return first error even with multiple failures', () => {
      const results: ValidationResult[] = [
        { isValid: false, error: 'Primeiro erro' },
        { isValid: false, error: 'Segundo erro' },
        { isValid: true }
      ];
      const combined = combineValidations(...results);
      expect(combined.isValid).toBe(false);
      expect(combined.error).toBe('Primeiro erro');
    });

    it('should handle single validation', () => {
      const result = combineValidations({ isValid: true });
      expect(result.isValid).toBe(true);
    });

    it('should handle single failed validation', () => {
      const result = combineValidations({ isValid: false, error: 'Falha' });
      expect(result.isValid).toBe(false);
      expect(result.error).toBe('Falha');
    });

    it('should work with real validation functions', () => {
      const result = combineValidations(
        validateRequired('John', 'Nome'),
        validateMinLength('John', 3, 'Nome'),
        validateMaxLength('John', 50, 'Nome')
      );
      expect(result.isValid).toBe(true);
    });

    it('should return first error from real validation functions', () => {
      const result = combineValidations(
        validateRequired('', 'Nome'),
        validateMinLength('', 3, 'Nome'),
        validateMaxLength('', 50, 'Nome')
      );
      expect(result.isValid).toBe(false);
      expect(result.error).toBe('Nome é obrigatório');
    });
  });
});
