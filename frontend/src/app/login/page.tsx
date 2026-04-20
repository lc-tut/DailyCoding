"use client";

import { useEffect, useState } from "react";
import { useSearchParams } from "next/navigation";
import { API_BASE } from "@/shared/config/api";

const PROVIDER_LABELS: Record<string, string> = {
  google: "Google",
  keycloak: "Keycloak",
};

const ERROR_MESSAGES: Record<string, { title: string; description: string }> = {
  access_denied: {
    title: "アクセス対象外のアカウントです",
    description: "このサービスはご利用いただけないアカウントです。対象のアカウントでお試しください。",
  },
  email_not_verified: {
    title: "メールアドレスが未確認です",
    description: "Googleアカウントのメールアドレスを確認してから再度お試しください。",
  },
};

export default function LoginPage() {
  const [providers, setProviders] = useState<string[]>([]);
  const [loaded, setLoaded] = useState(false);
  const searchParams = useSearchParams();
  const errorCode = searchParams.get("error");
  const errorInfo = errorCode ? ERROR_MESSAGES[errorCode] : null;
  const next = searchParams.get("next") ?? "/";

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

      {errorInfo ? (
        <div className="w-full max-w-xs rounded border border-red bg-surface0 p-4 flex flex-col gap-1">
          <p className="text-sm font-semibold text-red">{errorInfo.title}</p>
          <p className="text-xs text-subtext0">{errorInfo.description}</p>
        </div>
      ) : (
        <p className="text-sm text-subtext0">提出するにはログインが必要です</p>
      )}

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
              href={`${API_BASE}/api/auth/${provider}?next=${encodeURIComponent(next)}`}
              className="px-6 py-2 rounded bg-mauve text-base hover:opacity-80 transition-opacity font-semibold text-center"
            >
              {PROVIDER_LABELS[provider] ?? provider} でログイン
            </a>
          ))
        )}
      </div>
    </div>
  );
}
