all: format security test

security:
	@echo "*** running security checks... ***"
	@gosec ./...

format:
	@gofmt -s -w *.go _examples/**/*go

test:
	@go test -v -cover ./...

.PHONY: test
