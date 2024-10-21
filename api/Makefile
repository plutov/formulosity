AIR_BIN=bin/air

run:
	if [ ! -f $(AIR_BIN) ]; then \
		echo "Installing Air locally..."; \
		go build -o $(AIR_BIN) github.com/air-verse/air; \
	fi
	$(AIR_BIN)

lint:
	golangci-lint run

test:
	go test -v -bench=. -race ./...

gen: clean
	mockery --all --recursive --dir ./pkg --case underscore --with-expecter

clean:
	rm -rf mocks
