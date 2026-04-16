"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useUser } from "@/shared/providers/UserProvider";
import { API_BASE } from "@/shared/config/api";

export default function Header() {
  const user = useUser();
  const router = useRouter();

  const handleLogout = async () => {
    await fetch(`${API_BASE}/api/auth/logout`, {
      method: "POST",
      credentials: "include",
    });
    router.refresh();
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
          <a
            href={`${API_BASE}/api/auth/google`}
            className="text-xs px-3 py-1 rounded bg-[var(--color-mauve)] text-[var(--color-base)] hover:opacity-80 transition-opacity font-semibold"
          >
            Google でログイン
          </a>
        )}
      </div>
    </header>
  );
}
