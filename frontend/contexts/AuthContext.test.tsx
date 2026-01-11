import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import { renderHook, act } from '@testing-library/react';
import { AuthProvider, useAuth } from './AuthContext';
import { api } from '../services/api';
import type { User, AuthResponse, LoginRequest, RegisterRequest } from '../types';

// Mock the API service
vi.mock('../services/api', () => ({
  api: {
    getToken: vi.fn(),
    login: vi.fn(),
    register: vi.fn(),
    logout: vi.fn(),
  },
}));

describe('AuthContext', () => {
  const mockUser: User = {
    id: '1',
    name: 'Test User',
    email: 'test@example.com',
  };

  const mockAuthResponse: AuthResponse = {
    accessToken: 'mock-access-token-123',
    refreshToken: 'mock-refresh-token-123',
    expiresIn: 900,
    user: mockUser,
  };

  beforeEach(() => {
    // Clear all mocks before each test
    vi.clearAllMocks();
    localStorage.clear();
  });

  afterEach(() => {
    localStorage.clear();
  });

  describe('AuthProvider', () => {
    it('should render children', () => {
      render(
        <AuthProvider>
          <div>Test Content</div>
        </AuthProvider>
      );

      expect(screen.getByText('Test Content')).toBeDefined();
    });

    it('should initialize with loading state', () => {
      const { result } = renderHook(() => useAuth(), {
        wrapper: AuthProvider,
      });

      expect(result.current.loading).toBe(false); // becomes false after mount
    });

    it('should restore user from localStorage on mount', async () => {
      const storedUser = JSON.stringify(mockUser);
      localStorage.setItem('user', storedUser);
      vi.mocked(api.getToken).mockReturnValue('existing-token');

      const { result } = renderHook(() => useAuth(), {
        wrapper: AuthProvider,
      });

      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      expect(result.current.user).toEqual(mockUser);
      expect(result.current.isAuthenticated).toBe(true);
    });

    it('should initialize as unauthenticated when no token exists', async () => {
      vi.mocked(api.getToken).mockReturnValue(null);

      const { result } = renderHook(() => useAuth(), {
        wrapper: AuthProvider,
      });

      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      expect(result.current.user).toBeNull();
      expect(result.current.isAuthenticated).toBe(false);
    });

    it('should initialize as unauthenticated when token exists but no user in localStorage', async () => {
      vi.mocked(api.getToken).mockReturnValue('token-without-user');

      const { result } = renderHook(() => useAuth(), {
        wrapper: AuthProvider,
      });

      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      expect(result.current.user).toBeNull();
      expect(result.current.isAuthenticated).toBe(false);
    });
  });

  describe('login', () => {
    it('should successfully login user', async () => {
      vi.mocked(api.login).mockResolvedValue(mockAuthResponse);

      const { result } = renderHook(() => useAuth(), {
        wrapper: AuthProvider,
      });

      const loginData: LoginRequest = {
        email: 'test@example.com',
        password: 'password123',
      };

      await act(async () => {
        await result.current.login(loginData);
      });

      expect(api.login).toHaveBeenCalledWith(loginData);
      expect(result.current.user).toEqual(mockUser);
      expect(result.current.isAuthenticated).toBe(true);
      expect(localStorage.getItem('user')).toBe(JSON.stringify(mockUser));
    });

    it('should persist user data in localStorage after login', async () => {
      vi.mocked(api.login).mockResolvedValue(mockAuthResponse);

      const { result } = renderHook(() => useAuth(), {
        wrapper: AuthProvider,
      });

      await act(async () => {
        await result.current.login({
          email: 'test@example.com',
          password: 'password123',
        });
      });

      const storedUser = localStorage.getItem('user');
      expect(storedUser).toBe(JSON.stringify(mockUser));
    });

    it('should throw error on failed login', async () => {
      const errorMessage = 'Invalid credentials';
      vi.mocked(api.login).mockRejectedValue(new Error(errorMessage));

      const { result } = renderHook(() => useAuth(), {
        wrapper: AuthProvider,
      });

      await expect(async () => {
        await act(async () => {
          await result.current.login({
            email: 'wrong@example.com',
            password: 'wrongpassword',
          });
        });
      }).rejects.toThrow(errorMessage);

      expect(result.current.user).toBeNull();
      expect(result.current.isAuthenticated).toBe(false);
    });

    it('should handle multiple login attempts', async () => {
      const firstUser: User = { id: '1', name: 'User 1', email: 'user1@test.com' };
      const secondUser: User = { id: '2', name: 'User 2', email: 'user2@test.com' };

      vi.mocked(api.login)
        .mockResolvedValueOnce({ accessToken: 'token1', refreshToken: 'refresh1', expiresIn: 900, user: firstUser })
        .mockResolvedValueOnce({ accessToken: 'token2', refreshToken: 'refresh2', expiresIn: 900, user: secondUser });

      const { result } = renderHook(() => useAuth(), {
        wrapper: AuthProvider,
      });

      await act(async () => {
        await result.current.login({ email: 'user1@test.com', password: 'pass1' });
      });

      expect(result.current.user).toEqual(firstUser);

      await act(async () => {
        await result.current.login({ email: 'user2@test.com', password: 'pass2' });
      });

      expect(result.current.user).toEqual(secondUser);
    });
  });

  describe('register', () => {
    it('should successfully register user', async () => {
      vi.mocked(api.register).mockResolvedValue(mockAuthResponse);

      const { result } = renderHook(() => useAuth(), {
        wrapper: AuthProvider,
      });

      const registerData: RegisterRequest = {
        name: 'Test User',
        email: 'test@example.com',
        password: 'password123',
      };

      await act(async () => {
        await result.current.register(registerData);
      });

      expect(api.register).toHaveBeenCalledWith(registerData);
      expect(result.current.user).toEqual(mockUser);
      expect(result.current.isAuthenticated).toBe(true);
      expect(localStorage.getItem('user')).toBe(JSON.stringify(mockUser));
    });

    it('should persist user data in localStorage after registration', async () => {
      vi.mocked(api.register).mockResolvedValue(mockAuthResponse);

      const { result } = renderHook(() => useAuth(), {
        wrapper: AuthProvider,
      });

      await act(async () => {
        await result.current.register({
          name: 'New User',
          email: 'new@example.com',
          password: 'password123',
        });
      });

      const storedUser = localStorage.getItem('user');
      expect(storedUser).toBe(JSON.stringify(mockUser));
    });

    it('should throw error on failed registration', async () => {
      const errorMessage = 'Email already exists';
      vi.mocked(api.register).mockRejectedValue(new Error(errorMessage));

      const { result } = renderHook(() => useAuth(), {
        wrapper: AuthProvider,
      });

      await expect(async () => {
        await act(async () => {
          await result.current.register({
            name: 'Test User',
            email: 'existing@example.com',
            password: 'password123',
          });
        });
      }).rejects.toThrow(errorMessage);

      expect(result.current.user).toBeNull();
      expect(result.current.isAuthenticated).toBe(false);
    });

    it('should automatically authenticate user after registration', async () => {
      vi.mocked(api.register).mockResolvedValue(mockAuthResponse);

      const { result } = renderHook(() => useAuth(), {
        wrapper: AuthProvider,
      });

      expect(result.current.isAuthenticated).toBe(false);

      await act(async () => {
        await result.current.register({
          name: 'New User',
          email: 'new@example.com',
          password: 'password123',
        });
      });

      expect(result.current.isAuthenticated).toBe(true);
      expect(result.current.user).toEqual(mockUser);
    });
  });

  describe('logout', () => {
    it('should successfully logout user', async () => {
      // Setup: login first
      vi.mocked(api.login).mockResolvedValue(mockAuthResponse);

      const { result } = renderHook(() => useAuth(), {
        wrapper: AuthProvider,
      });

      await act(async () => {
        await result.current.login({
          email: 'test@example.com',
          password: 'password123',
        });
      });

      expect(result.current.isAuthenticated).toBe(true);

      // Action: logout
      act(() => {
        result.current.logout();
      });

      // Assert
      expect(api.logout).toHaveBeenCalled();
      expect(result.current.user).toBeNull();
      expect(result.current.isAuthenticated).toBe(false);
      expect(localStorage.getItem('user')).toBeNull();
    });

    it('should clear user from localStorage on logout', async () => {
      localStorage.setItem('user', JSON.stringify(mockUser));
      vi.mocked(api.getToken).mockReturnValue('token');

      const { result } = renderHook(() => useAuth(), {
        wrapper: AuthProvider,
      });

      await waitFor(() => {
        expect(result.current.user).toEqual(mockUser);
      });

      act(() => {
        result.current.logout();
      });

      expect(localStorage.getItem('user')).toBeNull();
    });

    it('should handle logout when not logged in', () => {
      const { result } = renderHook(() => useAuth(), {
        wrapper: AuthProvider,
      });

      act(() => {
        result.current.logout();
      });

      expect(result.current.user).toBeNull();
      expect(result.current.isAuthenticated).toBe(false);
      expect(api.logout).toHaveBeenCalled();
    });

    it('should allow login after logout', async () => {
      vi.mocked(api.login).mockResolvedValue(mockAuthResponse);

      const { result } = renderHook(() => useAuth(), {
        wrapper: AuthProvider,
      });

      // First login
      await act(async () => {
        await result.current.login({
          email: 'test@example.com',
          password: 'password123',
        });
      });

      expect(result.current.isAuthenticated).toBe(true);

      // Logout
      act(() => {
        result.current.logout();
      });

      expect(result.current.isAuthenticated).toBe(false);

      // Login again
      await act(async () => {
        await result.current.login({
          email: 'test@example.com',
          password: 'password123',
        });
      });

      expect(result.current.isAuthenticated).toBe(true);
    });
  });

  describe('isAuthenticated', () => {
    it('should return false when user is null', () => {
      const { result } = renderHook(() => useAuth(), {
        wrapper: AuthProvider,
      });

      expect(result.current.isAuthenticated).toBe(false);
    });

    it('should return true when user is set', async () => {
      vi.mocked(api.login).mockResolvedValue(mockAuthResponse);

      const { result } = renderHook(() => useAuth(), {
        wrapper: AuthProvider,
      });

      await act(async () => {
        await result.current.login({
          email: 'test@example.com',
          password: 'password123',
        });
      });

      expect(result.current.isAuthenticated).toBe(true);
    });

    it('should update when user state changes', async () => {
      vi.mocked(api.login).mockResolvedValue(mockAuthResponse);

      const { result } = renderHook(() => useAuth(), {
        wrapper: AuthProvider,
      });

      expect(result.current.isAuthenticated).toBe(false);

      await act(async () => {
        await result.current.login({
          email: 'test@example.com',
          password: 'password123',
        });
      });

      expect(result.current.isAuthenticated).toBe(true);

      act(() => {
        result.current.logout();
      });

      expect(result.current.isAuthenticated).toBe(false);
    });
  });

  describe('useAuth hook', () => {
    it('should throw error when used outside AuthProvider', () => {
      // Suppress console.error for this test
      const originalError = console.error;
      console.error = vi.fn();

      expect(() => {
        renderHook(() => useAuth());
      }).toThrow('useAuth must be used within an AuthProvider');

      console.error = originalError;
    });

    it('should return context value when used inside AuthProvider', () => {
      const { result } = renderHook(() => useAuth(), {
        wrapper: AuthProvider,
      });

      expect(result.current).toHaveProperty('user');
      expect(result.current).toHaveProperty('loading');
      expect(result.current).toHaveProperty('login');
      expect(result.current).toHaveProperty('register');
      expect(result.current).toHaveProperty('logout');
      expect(result.current).toHaveProperty('isAuthenticated');
    });
  });

  describe('Context Value Stability', () => {
    it('should memoize context value correctly', async () => {
      vi.mocked(api.login).mockResolvedValue(mockAuthResponse);

      const { result, rerender } = renderHook(() => useAuth(), {
        wrapper: AuthProvider,
      });

      const firstValue = result.current;

      // Rerender without changing state
      rerender();

      // Context value should be the same (memoized)
      expect(result.current).toBe(firstValue);
    });

    it('should update context value when user changes', async () => {
      vi.mocked(api.login).mockResolvedValue(mockAuthResponse);

      const { result } = renderHook(() => useAuth(), {
        wrapper: AuthProvider,
      });

      const initialValue = result.current;

      await act(async () => {
        await result.current.login({
          email: 'test@example.com',
          password: 'password123',
        });
      });

      // Context value should be different after state change
      expect(result.current).not.toBe(initialValue);
      expect(result.current.user).toEqual(mockUser);
    });
  });

  describe('Complete Authentication Flow', () => {
    it('should handle complete user journey: register -> logout -> login', async () => {
      const registerResponse: AuthResponse = {
        accessToken: 'register-token',
        refreshToken: 'register-refresh',
        expiresIn: 900,
        user: mockUser,
      };
      const loginResponse: AuthResponse = {
        accessToken: 'login-token',
        refreshToken: 'login-refresh',
        expiresIn: 900,
        user: mockUser,
      };

      vi.mocked(api.register).mockResolvedValue(registerResponse);
      vi.mocked(api.login).mockResolvedValue(loginResponse);

      const { result } = renderHook(() => useAuth(), {
        wrapper: AuthProvider,
      });

      // Initially not authenticated
      expect(result.current.isAuthenticated).toBe(false);

      // Register
      await act(async () => {
        await result.current.register({
          name: 'Test User',
          email: 'test@example.com',
          password: 'password123',
        });
      });

      expect(result.current.isAuthenticated).toBe(true);
      expect(result.current.user?.name).toBe('Test User');

      // Logout
      act(() => {
        result.current.logout();
      });

      expect(result.current.isAuthenticated).toBe(false);

      // Login
      await act(async () => {
        await result.current.login({
          email: 'test@example.com',
          password: 'password123',
        });
      });

      expect(result.current.isAuthenticated).toBe(true);
      expect(result.current.user?.email).toBe('test@example.com');
    });
  });
});
