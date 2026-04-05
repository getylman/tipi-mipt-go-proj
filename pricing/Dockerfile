FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.work go.work.sum* ./
COPY shared/ ./shared/
COPY pricing/ ./pricing/
WORKDIR /app/pricing
RUN go build -o /bin/pricing ./cmd/main.go

FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata wget
COPY --from=builder /bin/pricing /bin/pricing
COPY --from=builder /app/pricing/migrations/ /migrations/
EXPOSE 8081
ENTRYPOINT ["/bin/pricing"]
