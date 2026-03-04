# Build stage
FROM node:20-alpine AS builder

WORKDIR /app

# Install pnpm globally
RUN npm install -g pnpm

# Copy workspace root package files
COPY package.json ./

# Copy app-level package files and lock file
COPY apps/web/package.json ./apps/web/
COPY apps/web/pnpm-lock.yaml ./apps/web/

# Copy packages if any
COPY packages ./packages

# Copy turbo config
COPY turbo.json ./

# Install dependencies using the app-level lock file
WORKDIR /app/apps/web
RUN pnpm install --frozen-lockfile

# Back to workspace root
WORKDIR /app

# Copy source code
COPY apps/web ./apps/web

# Build the application
RUN pnpm --filter web build

# Runtime stage
FROM node:20-alpine

# Install pnpm
RUN npm install -g pnpm

WORKDIR /app

# Copy built application
COPY --from=builder /app/apps/web/.next ./apps/web/.next
COPY --from=builder /app/apps/web/public ./apps/web/public
COPY --from=builder /app/apps/web/package.json ./apps/web/
COPY --from=builder /app/apps/web/pnpm-lock.yaml ./apps/web/
COPY --from=builder /app/apps/web/node_modules ./apps/web/node_modules
COPY --from=builder /app/package.json ./

# Create non-root user
RUN addgroup -g 1001 -S nodejs && \
    adduser -S nextjs -u 1001

# Change ownership
RUN chown -R nextjs:nodejs /app

# Switch to non-root user
USER nextjs

# Expose port
EXPOSE 3000

# Set environment
ENV NODE_ENV=production
ENV PORT=3000

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:3000/ || exit 1

# Run the application
WORKDIR /app/apps/web
CMD ["pnpm", "start"]
