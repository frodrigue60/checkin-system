# Build stage
FROM oven/bun:latest AS builder

WORKDIR /app

# Copy lockfile and package.json
COPY package.json bun.lock ./

# Install dependencies
RUN bun install --frozen-lockfile

# Copy source code
COPY . .

# Build the application
# Note: For production Docker, it's recommended to use @sveltejs/adapter-node
RUN bun run build

# Run stage
FROM oven/bun:latest

WORKDIR /app

# Copy build artifacts and dependencies
# We copy node_modules because adapter-node needs them, or we can use a standalone build
COPY --from=builder /app/build ./build
COPY --from=builder /app/package.json .
COPY --from=builder /app/node_modules ./node_modules

# Environment variables
ENV NODE_ENV=production
ENV PORT=3000

# Expose port
EXPOSE 3000

# Run the app
CMD ["bun", "./build/index.js"]
