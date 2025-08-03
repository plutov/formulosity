
lint:
	golangci-lint run

test:
	go test -v -bench=. -race ./...

gen: clean
	mockery

clean:
	rm -rf mocks
