FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o warehouse-api

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/warehouse-api .
COPY --from=builder /app/sql/schema ./sql/schema
EXPOSE 8080
CMD [ "./warehouse-api" ]
