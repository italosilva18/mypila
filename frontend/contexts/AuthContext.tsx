import React, { createContext, useContext, useState, useEffect, useMemo, useCallback } from 'react';
import { User, LoginRequest, RegisterRequest } from '../types';
import { api } from '../services/api';

interface AuthContextType {
    user: User | null;
    loading: boolean;
    login: (data: LoginRequest) => Promise<void>;
    register: (data: RegisterRequest) => Promise<void>;
    logout: () => void;
    isAuthenticated: boolean;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
    const [user, setUser] = useState<User | null>(null);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        checkAuth();
    }, []);

    const checkAuth = async () => {
        const token = api.getToken();
        if (token) {
            // ideally we would validate token with backend, but for now we decode or just trust presence
            // In a real app, you might have a /me endpoint
            // For this demo, we can persist partial user info in localStorage or assume token valid
            // Let's decode if we had jwt-decode, or just set isAuthenticated.
            // For simplicity: if token exists, we set a temporary user object or fetch it.
            // Let's assume the user object is stored in localStorage too for persistence across refreshes
            const storedUser = localStorage.getItem('user');
            if (storedUser) {
                setUser(JSON.parse(storedUser));
            }
        }
        setLoading(false);
    };

    // Memoized login - prevents recreation on every render
    const login = useCallback(async (data: LoginRequest) => {
        const response = await api.login(data);
        setUser(response.user);
        localStorage.setItem('user', JSON.stringify(response.user));
    }, []);

    // Memoized register - prevents recreation on every render
    const register = useCallback(async (data: RegisterRequest) => {
        const response = await api.register(data);
        setUser(response.user);
        localStorage.setItem('user', JSON.stringify(response.user));
    }, []);

    // Memoized logout - prevents recreation on every render
    const logout = useCallback(() => {
        api.logout();
        setUser(null);
        localStorage.removeItem('user');
    }, []);

    const contextValue = useMemo(() => ({
        user,
        loading,
        login,
        register,
        logout,
        isAuthenticated: !!user
    }), [user, loading, login, register, logout]);

    return (
        <AuthContext.Provider value={contextValue}>
            {children}
        </AuthContext.Provider>
    );
};

export const useAuth = () => {
    const context = useContext(AuthContext);
    if (context === undefined) {
        throw new Error('useAuth must be used within an AuthProvider');
    }
    return context;
};
