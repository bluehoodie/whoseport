build:
	go build -o whoseport .

install: build
	cp whoseport $$GOPATH/bin/ && rm ./whoseport