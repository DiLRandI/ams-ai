import { useMemo, useState } from "react";
import { api, setToken } from "../../api/client";
import type { User } from "../../api/types";
import { AuthContext, type AuthContextValue } from "./AuthContext";

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(() => {
    const raw = localStorage.getItem("ams_user");
    return raw ? (JSON.parse(raw) as User) : null;
  });

  const value = useMemo<AuthContextValue>(
    () => ({
      user,
      bootstrapped: true,
      async login(email: string, password: string) {
        const response = await api.login(email, password);
        setToken(response.token);
        localStorage.setItem("ams_user", JSON.stringify(response.user));
        setUser(response.user);
      },
      logout() {
        setToken(null);
        localStorage.removeItem("ams_user");
        setUser(null);
      },
    }),
    [user],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}
