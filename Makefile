all:
	go build
	go vet
	go test

clean:
	go clean
