build:
	GO111MODULE=auto go build -o whoseport .

install: build
	cp whoseport $$GOPATH/bin/ && rm ./whoseport
