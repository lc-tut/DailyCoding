// SSR/SSC (サーバーサイド) は内部 Docker ネットワーク経由で直接バックエンドに接続する。
// BACKEND_INTERNAL_URL は NEXT_PUBLIC_ でないため JS バンドルに含まれず、
// サーバー上でのみ評価される。
//
// ブラウザ (CSR) は NEXT_PUBLIC_API_BASE を使う。
// 本番では同一オリジン (https://dailycoding.lcn.ad.jp) を指し、
// Caddy がリクエストをバックエンドに転送する。
export const API_BASE =
  typeof window === "undefined"
    ? (process.env.BACKEND_INTERNAL_URL ?? "http://localhost:8080")
    : (process.env.NEXT_PUBLIC_API_BASE ?? "http://localhost:8080");
