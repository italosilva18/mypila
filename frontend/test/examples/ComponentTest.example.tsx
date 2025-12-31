/**
 * EXAMPLE: Component Testing Template
 *
 * Este arquivo demonstra as melhores práticas para testar componentes React.
 * Use como referência ao criar novos testes de componentes.
 */

import { describe, it, expect, beforeEach, vi } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import React, { useState } from 'react';

// Example Component to Test
interface LoginFormProps {
  onSubmit: (email: string, password: string) => Promise<void>;
  onForgotPassword?: () => void;
  isLoading?: boolean;
}

const LoginForm: React.FC<LoginFormProps> = ({ onSubmit, onForgotPassword, isLoading }) => {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    if (!email || !password) {
      setError('Todos os campos são obrigatórios');
      return;
    }

    try {
      await onSubmit(email, password);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Erro ao fazer login');
    }
  };

  return (
    <form onSubmit={handleSubmit} aria-label="Login form">
      <h1>Login</h1>

      <div>
        <label htmlFor="email">Email</label>
        <input
          id="email"
          type="email"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          disabled={isLoading}
          placeholder="seu@email.com"
        />
      </div>

      <div>
        <label htmlFor="password">Senha</label>
        <input
          id="password"
          type="password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          disabled={isLoading}
          placeholder="••••••••"
        />
      </div>

      {error && (
        <div role="alert" aria-live="polite">
          {error}
        </div>
      )}

      <button type="submit" disabled={isLoading}>
        {isLoading ? 'Carregando...' : 'Entrar'}
      </button>

      {onForgotPassword && (
        <button type="button" onClick={onForgotPassword}>
          Esqueci minha senha
        </button>
      )}
    </form>
  );
};

// ============================================================================
// TESTS
// ============================================================================

describe('LoginForm Component', () => {
  // Mock functions
  const mockOnSubmit = vi.fn();
  const mockOnForgotPassword = vi.fn();

  beforeEach(() => {
    // Clear mocks before each test
    vi.clearAllMocks();
  });

  describe('Rendering', () => {
    it('should render all form elements', () => {
      render(<LoginForm onSubmit={mockOnSubmit} />);

      expect(screen.getByRole('heading', { name: /login/i })).toBeDefined();
      expect(screen.getByLabelText(/email/i)).toBeDefined();
      expect(screen.getByLabelText(/senha/i)).toBeDefined();
      expect(screen.getByRole('button', { name: /entrar/i })).toBeDefined();
    });

    it('should render forgot password button when callback provided', () => {
      render(<LoginForm onSubmit={mockOnSubmit} onForgotPassword={mockOnForgotPassword} />);

      expect(screen.getByRole('button', { name: /esqueci minha senha/i })).toBeDefined();
    });

    it('should not render forgot password button when callback not provided', () => {
      render(<LoginForm onSubmit={mockOnSubmit} />);

      expect(screen.queryByRole('button', { name: /esqueci minha senha/i })).toBeNull();
    });

    it('should have proper accessibility attributes', () => {
      render(<LoginForm onSubmit={mockOnSubmit} />);

      const form = screen.getByRole('form', { name: /login form/i });
      expect(form).toBeDefined();

      const emailInput = screen.getByLabelText(/email/i);
      expect(emailInput.getAttribute('type')).toBe('email');

      const passwordInput = screen.getByLabelText(/senha/i);
      expect(passwordInput.getAttribute('type')).toBe('password');
    });
  });

  describe('User Interactions', () => {
    it('should update email field on user input', async () => {
      const user = userEvent.setup();
      render(<LoginForm onSubmit={mockOnSubmit} />);

      const emailInput = screen.getByLabelText(/email/i) as HTMLInputElement;
      await user.type(emailInput, 'test@example.com');

      expect(emailInput.value).toBe('test@example.com');
    });

    it('should update password field on user input', async () => {
      const user = userEvent.setup();
      render(<LoginForm onSubmit={mockOnSubmit} />);

      const passwordInput = screen.getByLabelText(/senha/i) as HTMLInputElement;
      await user.type(passwordInput, 'password123');

      expect(passwordInput.value).toBe('password123');
    });

    it('should call onSubmit with correct values on form submission', async () => {
      const user = userEvent.setup();
      mockOnSubmit.mockResolvedValue(undefined);

      render(<LoginForm onSubmit={mockOnSubmit} />);

      await user.type(screen.getByLabelText(/email/i), 'test@example.com');
      await user.type(screen.getByLabelText(/senha/i), 'password123');
      await user.click(screen.getByRole('button', { name: /entrar/i }));

      expect(mockOnSubmit).toHaveBeenCalledWith('test@example.com', 'password123');
      expect(mockOnSubmit).toHaveBeenCalledTimes(1);
    });

    it('should call onForgotPassword when button is clicked', async () => {
      const user = userEvent.setup();
      render(<LoginForm onSubmit={mockOnSubmit} onForgotPassword={mockOnForgotPassword} />);

      await user.click(screen.getByRole('button', { name: /esqueci minha senha/i }));

      expect(mockOnForgotPassword).toHaveBeenCalledTimes(1);
    });
  });

  describe('Validation', () => {
    it('should show error when submitting empty form', async () => {
      const user = userEvent.setup();
      render(<LoginForm onSubmit={mockOnSubmit} />);

      await user.click(screen.getByRole('button', { name: /entrar/i }));

      expect(screen.getByRole('alert')).toHaveTextContent('Todos os campos são obrigatórios');
      expect(mockOnSubmit).not.toHaveBeenCalled();
    });

    it('should show error when email is empty', async () => {
      const user = userEvent.setup();
      render(<LoginForm onSubmit={mockOnSubmit} />);

      await user.type(screen.getByLabelText(/senha/i), 'password123');
      await user.click(screen.getByRole('button', { name: /entrar/i }));

      expect(screen.getByRole('alert')).toHaveTextContent('Todos os campos são obrigatórios');
      expect(mockOnSubmit).not.toHaveBeenCalled();
    });

    it('should show error when password is empty', async () => {
      const user = userEvent.setup();
      render(<LoginForm onSubmit={mockOnSubmit} />);

      await user.type(screen.getByLabelText(/email/i), 'test@example.com');
      await user.click(screen.getByRole('button', { name: /entrar/i }));

      expect(screen.getByRole('alert')).toHaveTextContent('Todos os campos são obrigatórios');
      expect(mockOnSubmit).not.toHaveBeenCalled();
    });

    it('should clear error when resubmitting with valid data', async () => {
      const user = userEvent.setup();
      mockOnSubmit.mockResolvedValue(undefined);

      render(<LoginForm onSubmit={mockOnSubmit} />);

      // Submit empty form to trigger error
      await user.click(screen.getByRole('button', { name: /entrar/i }));
      expect(screen.getByRole('alert')).toBeDefined();

      // Fill form and resubmit
      await user.type(screen.getByLabelText(/email/i), 'test@example.com');
      await user.type(screen.getByLabelText(/senha/i), 'password123');
      await user.click(screen.getByRole('button', { name: /entrar/i }));

      // Error should be cleared
      await waitFor(() => {
        expect(screen.queryByRole('alert')).toBeNull();
      });
    });
  });

  describe('Async Behavior', () => {
    it('should handle successful submission', async () => {
      const user = userEvent.setup();
      mockOnSubmit.mockResolvedValue(undefined);

      render(<LoginForm onSubmit={mockOnSubmit} />);

      await user.type(screen.getByLabelText(/email/i), 'test@example.com');
      await user.type(screen.getByLabelText(/senha/i), 'password123');
      await user.click(screen.getByRole('button', { name: /entrar/i }));

      await waitFor(() => {
        expect(mockOnSubmit).toHaveBeenCalled();
      });

      expect(screen.queryByRole('alert')).toBeNull();
    });

    it('should handle failed submission', async () => {
      const user = userEvent.setup();
      const errorMessage = 'Credenciais inválidas';
      mockOnSubmit.mockRejectedValue(new Error(errorMessage));

      render(<LoginForm onSubmit={mockOnSubmit} />);

      await user.type(screen.getByLabelText(/email/i), 'test@example.com');
      await user.type(screen.getByLabelText(/senha/i), 'wrongpassword');
      await user.click(screen.getByRole('button', { name: /entrar/i }));

      await waitFor(() => {
        expect(screen.getByRole('alert')).toHaveTextContent(errorMessage);
      });
    });
  });

  describe('Loading State', () => {
    it('should disable inputs when loading', () => {
      render(<LoginForm onSubmit={mockOnSubmit} isLoading={true} />);

      const emailInput = screen.getByLabelText(/email/i);
      const passwordInput = screen.getByLabelText(/senha/i);
      const submitButton = screen.getByRole('button', { name: /carregando/i });

      expect(emailInput).toHaveProperty('disabled', true);
      expect(passwordInput).toHaveProperty('disabled', true);
      expect(submitButton).toHaveProperty('disabled', true);
    });

    it('should show loading text on submit button', () => {
      render(<LoginForm onSubmit={mockOnSubmit} isLoading={true} />);

      expect(screen.getByRole('button', { name: /carregando/i })).toBeDefined();
      expect(screen.queryByRole('button', { name: /^entrar$/i })).toBeNull();
    });

    it('should enable inputs when not loading', () => {
      render(<LoginForm onSubmit={mockOnSubmit} isLoading={false} />);

      const emailInput = screen.getByLabelText(/email/i);
      const passwordInput = screen.getByLabelText(/senha/i);
      const submitButton = screen.getByRole('button', { name: /entrar/i });

      expect(emailInput).toHaveProperty('disabled', false);
      expect(passwordInput).toHaveProperty('disabled', false);
      expect(submitButton).toHaveProperty('disabled', false);
    });
  });

  describe('Edge Cases', () => {
    it('should handle rapid form submissions', async () => {
      const user = userEvent.setup();
      mockOnSubmit.mockResolvedValue(undefined);

      render(<LoginForm onSubmit={mockOnSubmit} />);

      await user.type(screen.getByLabelText(/email/i), 'test@example.com');
      await user.type(screen.getByLabelText(/senha/i), 'password123');

      const submitButton = screen.getByRole('button', { name: /entrar/i });
      await user.click(submitButton);
      await user.click(submitButton);
      await user.click(submitButton);

      // Should be called multiple times (no built-in debounce)
      await waitFor(() => {
        expect(mockOnSubmit.mock.calls.length).toBeGreaterThan(0);
      });
    });

    it('should handle special characters in input', async () => {
      const user = userEvent.setup();
      mockOnSubmit.mockResolvedValue(undefined);

      render(<LoginForm onSubmit={mockOnSubmit} />);

      const specialEmail = "test+tag@example.com";
      const specialPassword = "P@ssw0rd!#$%";

      await user.type(screen.getByLabelText(/email/i), specialEmail);
      await user.type(screen.getByLabelText(/senha/i), specialPassword);
      await user.click(screen.getByRole('button', { name: /entrar/i }));

      expect(mockOnSubmit).toHaveBeenCalledWith(specialEmail, specialPassword);
    });
  });
});

// ============================================================================
// PATTERNS & BEST PRACTICES DEMONSTRATED
// ============================================================================

/**
 * 1. QUERY PRIORITIES (in order of preference):
 *    - getByRole (accessibility-focused)
 *    - getByLabelText (forms)
 *    - getByPlaceholderText (forms)
 *    - getByText (content)
 *    - getByTestId (last resort)
 *
 * 2. USER-EVENT vs FIRE-EVENT:
 *    - Prefer userEvent.setup() for realistic user interactions
 *    - Use fireEvent only for low-level events
 *
 * 3. ASYNC TESTING:
 *    - Use waitFor() for async state changes
 *    - Use findBy* queries for async elements
 *    - Don't use act() manually (React Testing Library handles it)
 *
 * 4. ACCESSIBILITY:
 *    - Test with screen readers in mind
 *    - Use semantic HTML and ARIA attributes
 *    - Test keyboard navigation
 *
 * 5. TEST ORGANIZATION:
 *    - Group by feature/behavior
 *    - Use descriptive test names
 *    - Follow AAA pattern (Arrange-Act-Assert)
 *
 * 6. MOCKING:
 *    - Mock external dependencies
 *    - Clear mocks between tests
 *    - Verify mock calls and arguments
 *
 * 7. EDGE CASES:
 *    - Test loading states
 *    - Test error states
 *    - Test empty states
 *    - Test boundary values
 */
