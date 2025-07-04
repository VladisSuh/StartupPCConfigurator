import { createContext, useContext, ReactNode, useState, useCallback } from 'react';

interface AuthContextType {
    isAuthenticated: boolean;
    login: (token: string) => void;
    logout: () => void;
    getToken: () => string | null; 
}

const AuthContext = createContext<AuthContextType | null>(null);

export const AuthProvider = ({ children }: { children: ReactNode }) => {
    const [isAuthenticated, setIsAuthenticated] = useState(
        Boolean(localStorage.getItem('authToken'))
    );

    const login = useCallback((token: string) => {
        localStorage.setItem('authToken', token);
        console.log('Login successful. Token stored.', token);
        setIsAuthenticated(true);
        console.log('Login successful. Token stored.');
    }, []);

    const logout = useCallback(() => {
        localStorage.removeItem('authToken');
        setIsAuthenticated(false);
        console.log('Logged out. Token removed.');
    }, []);

    const getToken = useCallback(() => {
        return localStorage.getItem('authToken');
    }, []);

    const value = {
        isAuthenticated,
        login,
        logout,
        getToken 
    };

    return (
        <AuthContext.Provider value={value}>
            {children}
        </AuthContext.Provider>
    );
};

export const useAuth = () => {
    const context = useContext(AuthContext);
    if (!context) {
        throw new Error('useAuth must be used within AuthProvider');
    }
    return context;
};