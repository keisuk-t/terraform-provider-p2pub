
depends:
	dep ensure

build: depends
	go build -o terraform-provider-p2pub

test: depends
	go test ./...
