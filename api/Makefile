
lint:
	golangci-lint run

test:
	go test -v -bench=. -race ./...

gen: clean sqlc
	mockery

sqlc:
	sqlc generate

clean:
	rm -rf mocks
