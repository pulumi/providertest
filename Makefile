MAKEFLAGS    := --warn-undefined-variables

.PHONY: test
test:
	go test -v -coverprofile="coverage.txt" -coverpkg=./... ./...
