cli:
	go build -o ./bin/chat-cli main.go

test:
	go test ./... -v

test-coverage:
	go test ./... -v -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

test-short:
	go test ./... -short

benchmark:
	go test ./... -bench=.

clean-test:
	rm -f coverage.out coverage.html

lint:
	go vet ./...
	go fmt ./...

clean:
	rm -rf ./bin/
	rm -f coverage.out coverage.html

.PHONY: cli test test-coverage test-short benchmark clean-test lint clean