/**
 * Custom Test Utilities
 *
 * Helpers e utilities customizadas para facilitar a escrita de testes
 */

import React, { ReactElement } from 'react';
import { render, RenderOptions, RenderResult } from '@testing-library/react';
import { AuthProvider } from '../../contexts/AuthContext';
import type { User } from '../../types';

// ============================================================================
// Custom Render with Providers
// ============================================================================

interface AllProvidersProps {
  children: React.ReactNode;
}

/**
 * Wrapper com todos os providers da aplicacao
 */
const AllProviders: React.FC<AllProvidersProps> = ({ children }) => {
  return (
    <AuthProvider>
      {children}
    </AuthProvider>
  );
};

/**
 * Custom render que inclui todos os providers
 *
 * @example
 * ```typescript
 * import { renderWithProviders } from '@/test/helpers/testUtils';
 *
 * it('should render component', () => {
 *   renderWithProviders(<MyComponent />);
 *   expect(screen.getByText('Hello')).toBeDefined();
 * });
 * ```
 */
export function renderWithProviders(
  ui: ReactElement,
  options?: Omit<RenderOptions, 'wrapper'>
): RenderResult {
  return render(ui, { wrapper: AllProviders, ...options });
}

// ============================================================================
// Mock Data Factories
// ============================================================================

/**
 * Factory para criar usuarios mock
 */
export const createMockUser = (overrides?: Partial<User>): User => ({
  id: '1',
  name: 'Test User',
  email: 'test@example.com',
  ...overrides,
});

/**
 * Factory para criar multiplos usuarios
 */
export const createMockUsers = (count: number): User[] => {
  return Array.from({ length: count }, (_, i) => createMockUser({
    id: String(i + 1),
    name: `User ${i + 1}`,
    email: `user${i + 1}@example.com`,
  }));
};

// ============================================================================
// Custom Matchers
// ============================================================================

/**
 * Verifica se um elemento tem uma classe CSS
 */
export const toHaveClass = (element: HTMLElement, className: string): boolean => {
  return element.classList.contains(className);
};

/**
 * Verifica se um elemento esta visivel
 */
export const toBeVisible = (element: HTMLElement): boolean => {
  return element.style.display !== 'none' &&
         element.style.visibility !== 'hidden' &&
         element.style.opacity !== '0';
};

// ============================================================================
// Test Helpers
// ============================================================================

/**
 * Aguarda um determinado tempo (use com moderacao)
 *
 * @param ms - Tempo em milissegundos
 */
export const wait = (ms: number): Promise<void> => {
  return new Promise(resolve => setTimeout(resolve, ms));
};

/**
 * Simula digitacao lenta de usuario
 */
export const typeSlowly = async (
  element: HTMLElement,
  text: string,
  delay: number = 50
): Promise<void> => {
  for (const char of text) {
    element.dispatchEvent(new KeyboardEvent('keydown', { key: char }));
    if (element instanceof HTMLInputElement || element instanceof HTMLTextAreaElement) {
      element.value += char;
      element.dispatchEvent(new Event('input', { bubbles: true }));
    }
    await wait(delay);
  }
};

/**
 * Mock de localStorage para testes
 */
export class LocalStorageMock {
  private store: Map<string, string>;

  constructor() {
    this.store = new Map();
  }

  clear(): void {
    this.store.clear();
  }

  getItem(key: string): string | null {
    return this.store.get(key) || null;
  }

  setItem(key: string, value: string): void {
    this.store.set(key, value);
  }

  removeItem(key: string): void {
    this.store.delete(key);
  }

  get length(): number {
    return this.store.size;
  }

  key(index: number): string | null {
    return Array.from(this.store.keys())[index] || null;
  }
}

/**
 * Cria um mock de fetch com respostas customizadas
 */
export const createFetchMock = (responses: Record<string, any>) => {
  return vi.fn((url: string) => {
    const response = responses[url];
    if (!response) {
      return Promise.reject(new Error(`No mock response for ${url}`));
    }

    return Promise.resolve({
      ok: true,
      json: () => Promise.resolve(response),
      text: () => Promise.resolve(JSON.stringify(response)),
      status: 200,
      statusText: 'OK',
    });
  });
};

/**
 * Simula erro de rede
 */
export const createNetworkErrorMock = () => {
  return vi.fn(() => Promise.reject(new Error('Network error')));
};

/**
 * Simula resposta de API com delay
 */
export const createDelayedResponse = <T,>(data: T, delay: number = 1000) => {
  return vi.fn(() =>
    new Promise(resolve =>
      setTimeout(() => resolve({
        ok: true,
        json: () => Promise.resolve(data),
      }), delay)
    )
  );
};

// ============================================================================
// Query Helpers
// ============================================================================

/**
 * Encontra elemento por texto parcial
 */
export const findByTextContent = (
  container: HTMLElement,
  textMatch: string | RegExp
): HTMLElement | null => {
  return Array.from(container.querySelectorAll('*')).find(element => {
    const hasText = (node: Element): boolean => {
      if (typeof textMatch === 'string') {
        return node.textContent?.includes(textMatch) || false;
      }
      return textMatch.test(node.textContent || '');
    };
    return hasText(element);
  }) as HTMLElement || null;
};

/**
 * Verifica se elemento tem atributo data-testid
 */
export const getByTestId = (container: HTMLElement, testId: string): HTMLElement | null => {
  return container.querySelector(`[data-testid="${testId}"]`);
};

// ============================================================================
// Form Helpers
// ============================================================================

/**
 * Preenche formulario com valores
 */
export const fillForm = async (
  form: HTMLFormElement,
  values: Record<string, string>
): Promise<void> => {
  for (const [name, value] of Object.entries(values)) {
    const input = form.querySelector(`[name="${name}"]`) as HTMLInputElement;
    if (input) {
      input.value = value;
      input.dispatchEvent(new Event('input', { bubbles: true }));
      input.dispatchEvent(new Event('change', { bubbles: true }));
    }
  }
};

/**
 * Obtem valores de formulario
 */
export const getFormValues = (form: HTMLFormElement): Record<string, string> => {
  const formData = new FormData(form);
  const values: Record<string, string> = {};

  formData.forEach((value, key) => {
    values[key] = value.toString();
  });

  return values;
};

// ============================================================================
// Assertion Helpers
// ============================================================================

/**
 * Verifica se elemento tem atributos ARIA corretos
 */
export const hasAccessibleName = (element: HTMLElement, name: string): boolean => {
  return element.getAttribute('aria-label') === name ||
         element.getAttribute('aria-labelledby') !== null ||
         (element as HTMLInputElement).labels?.[0]?.textContent === name;
};

/**
 * Verifica se elemento e focavel
 */
export const isFocusable = (element: HTMLElement): boolean => {
  return element.tabIndex >= 0 || element.hasAttribute('tabindex');
};

/**
 * Verifica se elemento tem role ARIA
 */
export const hasRole = (element: HTMLElement, role: string): boolean => {
  return element.getAttribute('role') === role;
};

// ============================================================================
// Async Helpers
// ============================================================================

/**
 * Aguarda ate que uma condicao seja verdadeira
 *
 * @param condition - Funcao que retorna boolean
 * @param timeout - Timeout maximo em ms
 * @param interval - Intervalo de verificacao em ms
 */
export const waitForCondition = async (
  condition: () => boolean,
  timeout: number = 5000,
  interval: number = 50
): Promise<void> => {
  const startTime = Date.now();

  while (!condition()) {
    if (Date.now() - startTime > timeout) {
      throw new Error('Timeout waiting for condition');
    }
    await wait(interval);
  }
};

/**
 * Aguarda ate que elemento apareca no DOM
 */
export const waitForElement = async (
  selector: string,
  timeout: number = 5000
): Promise<HTMLElement> => {
  let element: HTMLElement | null = null;

  await waitForCondition(() => {
    element = document.querySelector(selector);
    return element !== null;
  }, timeout);

  return element!;
};

// ============================================================================
// Mock Service Workers (MSW) Helpers
// ============================================================================

/**
 * Cria handler de sucesso para MSW
 */
export const createSuccessHandler = <T,>(endpoint: string, data: T) => {
  return {
    url: endpoint,
    method: 'GET',
    response: {
      status: 200,
      body: data,
    },
  };
};

/**
 * Cria handler de erro para MSW
 */
export const createErrorHandler = (endpoint: string, status: number, message: string) => {
  return {
    url: endpoint,
    method: 'GET',
    response: {
      status,
      body: { error: message },
    },
  };
};

// ============================================================================
// Re-export everything from testing library
// ============================================================================

export * from '@testing-library/react';
export { default as userEvent } from '@testing-library/user-event';

// Re-export with custom render as default
export { renderWithProviders as render };
