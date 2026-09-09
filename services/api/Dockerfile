FROM golang:1.25-alpine AS deps
WORKDIR /src/services/api
COPY services/api/go.mod services/api/go.sum ./
RUN GOWORK=off go mod download

FROM deps AS source
WORKDIR /src
COPY go.work ./
COPY services/api ./services/api
COPY packages/core-go ./packages/core-go
COPY packages/modules-go ./packages/modules-go

FROM source AS test
WORKDIR /src/services/api
CMD ["go", "test", "./..."]

FROM source AS build
WORKDIR /src/services/api
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/openride-api ./cmd/api

FROM alpine:3.22 AS runtime
RUN apk add --no-cache ca-certificates \
    && addgroup -S openride \
    && adduser -S -G openride -u 10001 openride
WORKDIR /app
COPY --from=build /out/openride-api /app/openride-api
USER openride
EXPOSE 8080
ENTRYPOINT ["/app/openride-api"]
