FROM golang:latest AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/coverflow

FROM golang:latest AS runtime

WORKDIR /srv/app

ENV PORT=8080 \
    DB_PATH=/srv/app/data/coverflow.db \
    BASE_URL=http://localhost:8080 \
    FRONTEND_URL=http://localhost:3000

COPY --from=builder /app/bin/coverflow ./coverflow

RUN mkdir -p data storage temp

EXPOSE 8080

VOLUME ["/srv/app/data", "/srv/app/storage"]

ENTRYPOINT ["./coverflow"]
