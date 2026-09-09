FROM golang:1.25-alpine AS build

WORKDIR /src
COPY go.work ./
COPY packages/core-go ./packages/core-go
COPY packages/modules-go ./packages/modules-go
COPY services/marketplace ./services/marketplace

WORKDIR /src/services/marketplace
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/marketplace-service ./cmd/marketplace

FROM alpine:3.22 AS runtime
RUN apk add --no-cache ca-certificates \
    && addgroup -S openride \
    && adduser -S -G openride -u 10001 openride

WORKDIR /app
COPY --from=build /out/marketplace-service /app/marketplace-service
USER openride
EXPOSE 8090
ENTRYPOINT ["/app/marketplace-service"]
