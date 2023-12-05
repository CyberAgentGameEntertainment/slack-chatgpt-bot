FROM golang:1.21.3-alpine as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -o main .
WORKDIR /dist
RUN cp /app/main .

FROM alpine:latest
COPY --from=builder /dist/main /
CMD ["/main"]