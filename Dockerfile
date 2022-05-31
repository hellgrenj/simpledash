FROM golang:1.18-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download
COPY . .
RUN go build -o main .

FROM alpine:3.15 as runtime
COPY --from=builder ./app/main main
COPY --from=builder ./app/static static
CMD ["./main"]