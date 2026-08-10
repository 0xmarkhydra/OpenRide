FROM golang:1.25-alpine AS deps
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

FROM deps AS test
COPY . .
CMD ["go", "test", "./..."]

FROM deps AS build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/flashx-api ./cmd/api

FROM alpine:3.22 AS runtime
RUN apk add --no-cache ca-certificates \
    && addgroup -S flashx \
    && adduser -S -G flashx -u 10001 flashx
WORKDIR /app
COPY --from=build /out/flashx-api /app/flashx-api
USER flashx
EXPOSE 8080
ENTRYPOINT ["/app/flashx-api"]
