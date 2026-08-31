FROM golang:1.25.6-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /app/server ./cmd/

FROM alpine:3.19

COPY --from=builder /app/server /app/server
EXPOSE 8088
CMD ["/app/server"]
