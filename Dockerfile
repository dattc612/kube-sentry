FROM node:24-alpine AS frontend-builder
WORKDIR /app

COPY package.json yarn.lock ./
RUN yarn install --frozen-lockfile

COPY client ./client
COPY tsconfig.json ./
RUN yarn build

FROM golang:1.26-alpine AS backend-builder
WORKDIR /app

# Install dependencies
COPY go.mod go.sum ./
RUN go mod download

COPY . .
COPY --from=frontend-builder /app/static ./static

RUN CGO_ENABLED=0 GOOS=linux go build -o server server.go

FROM alpine:3.24.1
WORKDIR /opt/gtts

RUN apk --no-cache add ca-certificates

COPY --from=backend-builder /app/server /opt/gtts/server
COPY --from=backend-builder /app/static /opt/gtts/static
COPY config.yaml /opt/gtts/

EXPOSE 6120

CMD ["./server"]