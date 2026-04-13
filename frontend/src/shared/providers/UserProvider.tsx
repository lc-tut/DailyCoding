"use client";

import { createContext, useContext } from "react";
import type { User } from "../model/user";
import { MOCK_USER } from "../model/user";

const UserContext = createContext<User>(MOCK_USER);

export function UserProvider({ children }: { children: React.ReactNode }) {
  // Phase2: ここで Keycloak のトークンを検証してユーザーをセットする
  return (
    <UserContext.Provider value={MOCK_USER}>{children}</UserContext.Provider>
  );
}

export function useUser(): User {
  return useContext(UserContext);
}
