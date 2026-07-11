FROM golang:1.25-alpine AS build

WORKDIR /src

COPY apps/api/go.mod apps/api/go.sum ./
RUN go mod download

COPY apps/api ./
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/linora-api ./cmd/server

FROM alpine:3.22

RUN addgroup -S linora && adduser -S linora -G linora
COPY --from=build /out/linora-api /usr/local/bin/linora-api

USER linora
EXPOSE 8080
ENTRYPOINT ["linora-api"]
