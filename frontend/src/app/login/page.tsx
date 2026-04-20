"use client";

import { useEffect, useState } from "react";
import { API_BASE } from "@/shared/config/api";

const PROVIDER_LABELS: Record<string, string> = {
  google: "Google",
  keycloak: "Keycloak",
};

export default function LoginPage() {
  const [providers, setProviders] = useState<string[]>([]);
  const [loaded, setLoaded] = useState(false);

  useEffect(() => {
    fetch(`${API_BASE}/api/auth/providers`, { credentials: "include" })
      .then((res) => (res.ok ? res.json() : { providers: [] }))
      .then((data) => setProviders(data.providers ?? []))
      .catch(() => setProviders([]))
      .finally(() => setLoaded(true));
  }, []);

  return (
    <div className="flex flex-col items-center justify-center flex-1 w-full gap-6 p-8">
      <h1 className="text-xl font-bold text-[var(--color-text)]">ログイン</h1>
      <p className="text-sm text-[var(--color-subtext0)]">提出するにはログインが必要です</p>
      <div className="flex flex-col gap-3 w-full max-w-xs">
        {!loaded ? (
          <p className="text-sm text-center text-[var(--color-overlay0)]">読み込み中...</p>
        ) : providers.length === 0 ? (
          <p className="text-sm text-center text-[var(--color-overlay0)]">
            ログイン方法が設定されていません
          </p>
        ) : (
          providers.map((provider) => (
            <a
              key={provider}
              href={`${API_BASE}/api/auth/${provider}`}
              className="px-6 py-2 rounded bg-[var(--color-mauve)] text-[var(--color-base)] hover:opacity-80 transition-opacity font-semibold text-center"
            >
              {PROVIDER_LABELS[provider] ?? provider} でログイン
            </a>
          ))
        )}
      </div>
    </div>
  );
}
