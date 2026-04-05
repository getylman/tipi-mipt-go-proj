FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.work go.work.sum* ./
COPY shared/ ./shared/
COPY ingestion/ ./ingestion/
WORKDIR /app/ingestion
RUN go build -o /bin/ingestion ./cmd/main.go

FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata wget
COPY --from=builder /bin/ingestion /bin/ingestion
COPY --from=builder /app/ingestion/migrations/ /migrations/
EXPOSE 8080
ENTRYPOINT ["/bin/ingestion"]
