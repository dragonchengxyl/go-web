# Build stage
FROM node:20-alpine AS builder

WORKDIR /app

ARG INTERNAL_API_ORIGIN=http://localhost:8080
ARG INTERNAL_WS_ORIGIN=http://localhost:8080

ENV NEXT_TELEMETRY_DISABLED=1
ENV INTERNAL_API_ORIGIN=${INTERNAL_API_ORIGIN}
ENV INTERNAL_WS_ORIGIN=${INTERNAL_WS_ORIGIN}

RUN npm install -g pnpm

COPY pnpm-workspace.yaml pnpm-lock.yaml package.json ./
COPY apps/web/package.json ./apps/web/
COPY packages/api-client/package.json ./packages/api-client/

RUN pnpm install --frozen-lockfile

COPY apps/web ./apps/web
COPY packages ./packages
COPY turbo.json ./

RUN pnpm --filter web build

# Runtime stage
FROM node:20-alpine

WORKDIR /app

ENV NODE_ENV=production
ENV PORT=3000
ENV HOSTNAME=0.0.0.0
ENV NEXT_TELEMETRY_DISABLED=1

COPY --from=builder /app/apps/web/.next/standalone ./
COPY --from=builder /app/apps/web/.next/static ./apps/web/.next/static
COPY --from=builder /app/apps/web/public ./apps/web/public

RUN addgroup -g 1001 -S nodejs && \
    adduser -S nextjs -u 1001 && \
    chown -R nextjs:nodejs /app

USER nextjs

EXPOSE 3000

HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:3000/ || exit 1

CMD ["node", "apps/web/server.js"]
