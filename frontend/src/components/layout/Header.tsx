"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { useUser } from "@/shared/providers/UserProvider";
import { API_BASE } from "@/shared/config/api";

const PROVIDER_LABELS: Record<string, string> = {
  google: "Google",
  keycloak: "Keycloak",
};

export default function Header() {
  const user = useUser();
  const [providers, setProviders] = useState<string[]>([]);

  useEffect(() => {
    fetch(`${API_BASE}/api/auth/providers`, { credentials: "include" })
      .then((res) => (res.ok ? res.json() : { providers: [] }))
      .then((data) => setProviders(data.providers ?? []))
      .catch(() => setProviders([]));
  }, []);

  const handleLogout = async () => {
    await fetch(`${API_BASE}/api/auth/logout`, {
      method: "POST",
      credentials: "include",
    });
    window.location.href = "/";
  };

  return (
    <header className="flex items-center gap-4 px-5 py-2.5 bg-[var(--color-mantle)] border-b border-[var(--color-surface0)] shrink-0">
      <Link
        href="/"
        className="text-lg font-bold text-[var(--color-mauve)] hover:text-[var(--color-lavender)] transition-colors"
      >
        DaileyCoding
      </Link>

      <nav className="flex items-center gap-4 ml-2">
        <Link
          href="/problems"
          className="text-sm text-[var(--color-subtext0)] hover:text-[var(--color-text)] transition-colors"
        >
          問題一覧
        </Link>

        {user && (
          <Link
            href="/problems/new"
            className="text-sm text-[var(--color-subtext0)] hover:text-[var(--color-text)] transition-colors"
          >
            問題を投稿
          </Link>
        )}

        {user?.role === "admin" && (
          <>
            <span className="text-[var(--color-surface1)]">|</span>
            <Link
              href="/admin/problems"
              className="text-sm text-[var(--color-peach)] hover:text-[var(--color-yellow)] transition-colors"
            >
              問題管理
            </Link>
          </>
        )}
      </nav>

      <div className="ml-auto flex items-center gap-2">
        {user ? (
          <>
            <span className="text-xs text-[var(--color-overlay0)]">{user.name}</span>
            {user.role === "admin" && (
              <span className="text-xs px-1.5 py-0.5 rounded bg-[color-mix(in_srgb,var(--color-peach)_20%,transparent)] text-[var(--color-peach)] border border-[var(--color-peach)] font-semibold">
                admin
              </span>
            )}
            <button
              onClick={handleLogout}
              className="text-xs text-[var(--color-subtext0)] hover:text-[var(--color-text)] transition-colors ml-2"
            >
              ログアウト
            </button>
          </>
        ) : (
          providers.map((provider) => (
            <a
              key={provider}
              href={`${API_BASE}/api/auth/${provider}`}
              className="text-xs px-3 py-1 rounded bg-[var(--color-mauve)] text-[var(--color-base)] hover:opacity-80 transition-opacity font-semibold"
            >
              {PROVIDER_LABELS[provider] ?? provider} でログイン
            </a>
          ))
        )}
      </div>
    </header>
  );
}
