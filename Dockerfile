FROM golang:latest AS builder
WORKDIR /app
ADD . /app/
RUN go mod download && go mod verify
RUN go build -o ./out/indexer ./cmd/indexer

FROM golang:latest
WORKDIR /app

COPY --from=builder /app/out/indexer ./out/
# COPY config.json ./
ENTRYPOINT ["/app/out/indexer"]