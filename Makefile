default:
	go vet ./...
	staticcheck ./...
	go test
	go build ./cmd/html-lint

clean:
	-rm -f html-lint
	go clean
