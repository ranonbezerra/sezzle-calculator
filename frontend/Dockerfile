# syntax=docker/dockerfile:1

# Must match .nvmrc; the build fails if they drift.
ARG NODE_VERSION=22

FROM node:${NODE_VERSION}-alpine AS build
WORKDIR /app
COPY .nvmrc ./
RUN test "$(node -p 'process.versions.node.split(".")[0]')" = "$(cat .nvmrc)" \
    || { echo "Node $(node -v) does not match .nvmrc ($(cat .nvmrc))" >&2; exit 1; }
COPY package.json package-lock.json ./
RUN --mount=type=cache,target=/root/.npm npm ci
COPY . .
RUN npm run build

FROM nginxinc/nginx-unprivileged:1.30.6-alpine
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /app/dist /usr/share/nginx/html
EXPOSE 8080
