import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from "react";
import { clearSession, loadSession, saveSession } from "../api/session";
import type { UserProfile } from "../api/auth";

type AuthContextValue = {
  user: UserProfile | null;
  isAuthenticated: boolean;
  setSession: (token: string, user: UserProfile) => void;
  signOut: () => void;
};

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<UserProfile | null>(() => loadSession()?.user ?? null);

  const setSession = useCallback((token: string, nextUser: UserProfile) => {
    saveSession(token, nextUser);
    setUser(nextUser);
  }, []);

  const signOut = useCallback(() => {
    clearSession();
    setUser(null);
  }, []);

  useEffect(() => {
    function onSessionCleared() {
      setUser(null);
    }
    window.addEventListener("internly:session-cleared", onSessionCleared);
    return () => window.removeEventListener("internly:session-cleared", onSessionCleared);
  }, []);

  const value = useMemo(
    () => ({
      user,
      isAuthenticated: user !== null,
      setSession,
      signOut,
    }),
    [user, setSession, signOut],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error("useAuth must be used within AuthProvider");
  }
  return context;
}
