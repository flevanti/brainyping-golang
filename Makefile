BINARY_FOLDER="./bin"
TIMESTAMP=$(shell date +%s)
GITHASH=$(shell git rev-parse --short HEAD)

build:
	rm -f ${BINARY_FOLDER}/darwin_arm64/* && GOOS=darwin GOARCH=arm64 go build -ldflags="-X 'brainyping/pkg/initapp.buildDateUnix=${TIMESTAMP}' -X 'brainyping/pkg/initapp.gitHash=${GITHASH}' -s -w"  -o ${BINARY_FOLDER}/darwin_arm64/ -v ./...
	rm -f ${BINARY_FOLDER}/linux_amd64/*  && GOOS=linux  GOARCH=amd64 go build -ldflags="-X 'brainyping/pkg/initapp.buildDateUnix=${TIMESTAMP}' -X 'brainyping/pkg/initapp.gitHash=${GITHASH}' -s -w"  -o ${BINARY_FOLDER}/linux_amd64/ -v ./...
	rm -f ${BINARY_FOLDER}/linux_arm64/*  && GOOS=linux  GOARCH=arm64 go build -ldflags="-X 'brainyping/pkg/initapp.buildDateUnix=${TIMESTAMP}' -X 'brainyping/pkg/initapp.gitHash=${GITHASH}' -s -w"  -o ${BINARY_FOLDER}/linux_arm64/ -v ./...

migration:
	go run ./cmd/migrate/
