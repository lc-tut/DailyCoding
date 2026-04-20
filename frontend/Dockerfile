# ── Stage 1: Dependencies ────────────────────────────────────────
FROM node:22-alpine AS deps

RUN corepack enable
WORKDIR /app
COPY package.json pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile

# ── Stage 2: Build ───────────────────────────────────────────────
FROM node:22-alpine AS builder

RUN corepack enable
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY . .

# NEXT_PUBLIC_ 変数はビルド時に展開される
# 本番ドメインに合わせて ARG で上書き可能
ARG NEXT_PUBLIC_API_BASE=https://dailycoding.lcn.ad.jp
ENV NEXT_PUBLIC_API_BASE=$NEXT_PUBLIC_API_BASE

RUN pnpm build

# ── Stage 3: Runtime ─────────────────────────────────────────────
FROM node:22-alpine AS runner

WORKDIR /app
ENV NODE_ENV=production
ENV PORT=3000
ENV HOSTNAME=0.0.0.0

# standalone に含まれない静的ファイルをコピー
COPY --from=builder /app/public ./public
COPY --from=builder /app/.next/standalone ./
COPY --from=builder /app/.next/static ./.next/static

EXPOSE 3000
CMD ["node", "server.js"]
