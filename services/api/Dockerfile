FROM golang:1.27-alpine AS deps
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

FROM deps AS test
COPY . .
CMD ["go", "test", "./..."]

FROM deps AS build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/openride-compat-api ./cmd/api

FROM alpine:3.22 AS runtime
RUN apk add --no-cache ca-certificates \
    && addgroup -S openride \
    && adduser -S -G openride -u 10001 openride
WORKDIR /app
COPY --from=build /out/openride-compat-api /app/openride-compat-api
USER openride
EXPOSE 8080
ENTRYPOINT ["/app/openride-compat-api"]
