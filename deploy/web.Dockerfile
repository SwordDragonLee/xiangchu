# 管理后台构建镜像（build context: 仓库根目录）
FROM node:20-alpine AS build
WORKDIR /app
RUN corepack enable
COPY web/pnpm-workspace.yaml web/.npmrc web/package.json ./
COPY web/packages ./packages
COPY web/apps/admin ./apps/admin
RUN pnpm install --no-frozen-lockfile && pnpm --filter admin build

FROM nginx:1.27-alpine
COPY --from=build /app/apps/admin/dist /usr/share/nginx/html
COPY deploy/nginx.conf /etc/nginx/conf.d/default.conf
