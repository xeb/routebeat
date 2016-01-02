BUILD_DIR=build
COVERAGE_DIR=${BUILD_DIR}/coverage

# Runs test suite as root
# See: github.com/aeden/traceroute

build:
	go get ./...
	go build

test:
	sudo env GOPATH=`echo ${GOPATH}` go test

coverage-report:
	mkdir -p ${COVERAGE_DIR}
	sudo env GOPATH=`echo ${GOPATH}` go test -coverprofile=${COVERAGE_DIR}/routebeat.cov
	go tool cover -html=${COVERAGE_DIR}/routebeat.cov -o ${COVERAGE_DIR}/routebeat.html

coverage:
	sudo env GOPATH=`echo ${GOPATH}` go test -cover

.PHONY: build
