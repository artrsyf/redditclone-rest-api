FROM golang:alpine AS builder
LABEL stage=gobuilder

ENV CGO_ENABLED 0
ENV GOOS linux

RUN apk update --no-cache && apk add --no-cache tzdata

WORKDIR /build

COPY ./go.mod .
COPY ./go.sum .

RUN go mod download

COPY . .

RUN go build -ldflags="-s -w" -o /app/redditclone ./cmd/redditclone/

FROM alpine

RUN apk update --no-cache

WORKDIR /app/hw6/bin/hw6/

COPY --from=builder /build/static/ /app/hw6/static/
COPY --from=builder /build/cmd/redditclone/.env /app/hw6/bin/hw6/.env
COPY --from=builder /build/cmd/redditclone/config.yaml /app/hw6/bin/hw6/config.yaml
COPY --from=builder /app/redditclone /app/hw6/bin/hw6/redditclone

EXPOSE 8080

CMD ["./redditclone"]