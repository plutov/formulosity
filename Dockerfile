FROM golang:1.22 as builder

WORKDIR /go/src/github.com/plutov/formulosity

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY ./cmd ./cmd
COPY ./pkg ./pkg
COPY ./migrations ./migrations
RUN CGO_ENABLED=0 GOOS=linux go build -o api cmd/console-api/api.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata bash
WORKDIR /root
COPY --from=builder /go/src/github.com/plutov/formulosity/api .
COPY --from=builder /go/src/github.com/plutov/formulosity/migrations ./migrations

CMD ["./api"]
