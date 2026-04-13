"use client";

import Link from "next/link";
import { useUser } from "@/shared/providers/UserProvider";

export default function Header() {
  const user = useUser();

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

        {/* 問題作成は全ユーザーが利用可能 */}
        <Link
          href="/problems/new"
          className="text-sm text-[var(--color-subtext0)] hover:text-[var(--color-text)] transition-colors"
        >
          問題を投稿
        </Link>

        {/* 問題管理は admin のみ */}
        {user.role === "admin" && (
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
        <span className="text-xs text-[var(--color-overlay0)]">{user.name}</span>
        {user.role === "admin" && (
          <span className="text-xs px-1.5 py-0.5 rounded bg-[color-mix(in_srgb,var(--color-peach)_20%,transparent)] text-[var(--color-peach)] border border-[var(--color-peach)] font-semibold">
            admin
          </span>
        )}
      </div>
    </header>
  );
}
