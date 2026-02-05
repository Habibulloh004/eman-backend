FROM golang:1.25-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN apk add --no-cache build-base libwebp-dev

RUN CGO_ENABLED=1 GOOS=linux go build -o /out/server ./

FROM alpine:3.20

RUN adduser -D -g '' app \
    && apk add --no-cache ca-certificates libwebp

WORKDIR /app

COPY --from=build /out/server /app/server

RUN mkdir -p /app/uploads \
    && chown -R app:app /app

USER app

ENV PORT=8080

EXPOSE 8080

CMD ["/app/server"]
