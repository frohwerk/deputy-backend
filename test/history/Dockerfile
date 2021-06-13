FROM golang:1.16
WORKDIR /go/src/test
COPY go.mod go.sum history_test.go ./
ENTRYPOINT [ "go", "test", "-v", "." ]