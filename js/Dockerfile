# syntax=docker/dockerfile:1

# Single Node stage: build the app with Rspack (via Rsbuild), then serve the
# build with Rsbuild's own preview server. No nginx — one toolchain.
FROM node:26-alpine
# node:26-alpine no longer bundles corepack, so install pnpm directly.
RUN npm install -g pnpm@10
WORKDIR /app
COPY package.json pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile
COPY . .
RUN pnpm build
# Where the /api proxy sends requests; override when goapi runs elsewhere.
ENV GOAPI_URL=http://goapi:8080
EXPOSE 3000
CMD ["pnpm", "preview"]
