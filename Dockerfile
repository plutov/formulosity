# Build API
FROM golang:1.22-alpine AS api_builder

RUN apk add build-base

WORKDIR /api

COPY --from=api go.mod .
COPY --from=api go.sum .

RUN go mod download

COPY --from=api ./cmd ./cmd
COPY --from=api ./pkg ./pkg
COPY --from=api ./migrations ./migrations
COPY --from=api ./surveys ./surveys-examples
RUN CGO_ENABLED=1 GOOS=linux go build -o api -tags enablecgo cmd/console-api/api.go


# Build UI
FROM node:20-alpine AS ui_base

FROM ui_base AS deps
RUN apk add --no-cache libc6-compat
WORKDIR /app

COPY --from=ui package.json package-lock.json ./
RUN npm ci

FROM ui_base AS ui_builder
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY --from=ui . .

ENV NODE_ENV=production

RUN npm run build


# Final image
FROM nginxinc/nginx-unprivileged:alpine AS runner

USER 0:0

RUN apk --no-cache add ca-certificates tzdata nodejs

WORKDIR /app
ENV NODE_ENV=production

COPY --from=ui_builder /app/public ./public

RUN mkdir .next
RUN chown 1000:1000 .next

COPY --from=ui_builder --chown=1000:1000 /app/.next/standalone ./
COPY --from=ui_builder --chown=1000:1000 /app/.next/static ./.next/static

WORKDIR /api

COPY --from=api_builder --chown=1000:1000 /api/api ./api
COPY --from=api_builder --chown=1000:1000 /api/migrations ./migrations
COPY --from=api_builder --chown=1000:1000 /api/surveys-examples ./surveys-examples

RUN mkdir /data
RUN chown 1000:1000 /data

COPY start.sh /start.sh
RUN chmod +x /start.sh

COPY nginx/default.conf /etc/nginx/conf.d/default.conf
RUN chown 1000:1000 /etc/nginx/conf.d/default.conf

USER 1000:1000
RUN mkdir /data/surveys
RUN mkdir /data/db

EXPOSE 8081

CMD ["sh", "/start.sh"]
