FROM node:20-slim AS build

WORKDIR /app

COPY package.json pnpm-lock.yaml ./
RUN corepack enable && corepack prepare pnpm@latest --activate
RUN pnpm install --frozen-lockfile

COPY . .

ARG APP_API_HOST
ARG APP_GITHUB_APP_NAME
ENV APP_API_HOST=${APP_API_HOST}
ENV APP_GITHUB_APP_NAME=${APP_GITHUB_APP_NAME}

RUN pnpm build

FROM nginx:alpine AS runtime

COPY --from=build /app/dist /usr/share/nginx/html

RUN echo 'server { \
    listen 80; \
    location / { \
    root /usr/share/nginx/html; \
    index index.html; \
    try_files $uri $uri/ /index.html; \
    } \
    }' > /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
