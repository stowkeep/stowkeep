# syntax=docker/dockerfile:1

FROM node:22-alpine AS web-build
WORKDIR /src/web
COPY web/package.json web/package-lock.json* ./
RUN npm ci --ignore-scripts 2>/dev/null || npm install
COPY web/ ./
RUN npm run build

FROM golang:1.25-alpine AS go-build
RUN apk add --no-cache git ca-certificates
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-s -w" -o /out/stowkeep ./cmd/stowkeep

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=go-build /out/stowkeep /stowkeep
COPY --from=go-build /src/migrations /migrations
ENV STOWKEEP_MIGRATIONS_DIR=/migrations
WORKDIR /
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/stowkeep"]
