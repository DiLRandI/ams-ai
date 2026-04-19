import { createContext, useContext } from 'react';
import type { User } from '../../api/types';

export interface AuthContextValue {
  user: User | null;
  bootstrapped: boolean;
  login(email: string, password: string): Promise<void>;
  logout(): void;
}

export const AuthContext = createContext<AuthContextValue | null>(null);

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) {
    throw new Error('useAuth must be used inside AuthProvider');
  }
  return ctx;
}
