# syntax=docker/dockerfile:1

# ---- Build stage ----
FROM node:22-alpine AS build
WORKDIR /app

# Install dependencies against the lockfile for reproducible builds.
COPY package.json package-lock.json ./
RUN npm ci

# Build the production bundle. VITE_-prefixed args are baked into the client,
# so keep secrets out of them. Defaults to "/api" (nginx proxies it below).
ARG VITE_API_BASE_URL=/api
ENV VITE_API_BASE_URL=$VITE_API_BASE_URL

COPY . .
RUN npm run build

# ---- Runtime stage ----
FROM nginx:1.27-alpine AS runtime

# SPA + /api reverse-proxy config, templated so nginx listens on $FRONTEND_PORT.
# The nginx image runs envsubst over /etc/nginx/templates/*.template at startup.
COPY nginx.conf.template /etc/nginx/templates/default.conf.template
COPY --from=build /app/dist /usr/share/nginx/html

# Port nginx listens on inside the container (overridable at runtime).
ENV FRONTEND_PORT=5173
# Only substitute our own var; leave nginx's $host/$uri/etc. untouched.
ENV NGINX_ENVSUBST_FILTER=FRONTEND_PORT

EXPOSE 5173
