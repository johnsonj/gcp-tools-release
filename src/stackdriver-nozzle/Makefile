default: test

# Builds stackdriver-nozzle for linux-amd64
build:
	go build -o out/stackdriver-nozzle .

# Build cross-platform binaries
build-all:
	gox -output="out/stackdriver-nozzle_{{.OS}}_{{.Arch}}" .

# Prepration for tests
get-deps:
	# Go lint tool
	go get github.com/golang/lint/golint

	# Simplify cross-compiling
	go get github.com/mitchellh/gox

	# Ginkgo and omega test tools
	go get github.com/onsi/ginkgo/ginkgo
	go get github.com/onsi/gomega

# Cleans up directory and source code with gofmt
clean:
	go clean ./...

# Run gofmt on all code
fmt:
	gofmt -l -w $$(ls -d */ | grep -v vendor) 

# Run linter with non-stric checking
lint:
	@echo ls -d */ | grep -v vendor | xargs -L 1 golint
	ls -d */ | grep -v vendor | xargs -L 1 golint

# Vet code
vet:
	go tool vet $$(ls -d */ | grep -v vendor)

# Runs the unit tests with coverage
test: get-deps clean fmt lint vet build
	ginkgo -r -skipPackage=integration .

newb: get-deps
	go get -u github.com/kardianos/govendor
