FROM golang:1.22 as builder

WORKDIR /root

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY ./cmd ./cmd
COPY ./pkg ./pkg
COPY ./migrations ./migrations
COPY ./surveys ./surveys-examples
RUN CGO_ENABLED=0 GOOS=linux go build -o api cmd/console-api/api.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata bash
WORKDIR /root
COPY --from=builder /root/api .
COPY --from=builder /root/migrations ./migrations
COPY --from=builder /root/surveys-examples ./surveys-examples

CMD ["./api"]
