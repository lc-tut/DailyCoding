"use client";

import { createContext, useContext, useEffect, useState } from "react";
import type { User } from "../model/user";
import { API_BASE } from "../config/api";

type UserState = User | null;

const UserContext = createContext<UserState>(null);

export function UserProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<UserState>(null);

  useEffect(() => {
    fetch(`${API_BASE}/api/auth/me`, { credentials: "include" })
      .then((res) => (res.ok ? res.json() : null))
      .then((data: User | null) => setUser(data))
      .catch(() => setUser(null));
  }, []);

  return <UserContext.Provider value={user}>{children}</UserContext.Provider>;
}

export function useUser(): UserState {
  return useContext(UserContext);
}
