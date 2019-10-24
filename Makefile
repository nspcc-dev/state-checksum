# All of the targets are phony here because we don't really use make dependency
# tracking for files
.PHONY: build deps image check-version clean-cluster push-tag push-to-registry \
	run run-cluster test vet lint fmt cover

test:
	@go test ./... -cover

vet:
	@go vet ./...

lint:
	@go list ./... | xargs -L1 golint -set_exit_status

fmt:
	@gofmt -l -w -s $$(find . -type f -name '*.go'| grep -v "/vendor/")

cover:
	@go test -v -race ./... -coverprofile=coverage.txt -covermode=atomic
	@go tool cover -html=coverage.txt -o coverage.html
