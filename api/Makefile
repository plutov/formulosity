lint:
	golangci-lint run

test:
	go test -v -bench=. -race ./...

gen: clean
	mockery --all --recursive --dir ./pkg --case underscore --with-expecter

clean:
	rm -rf mocks
