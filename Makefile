

MODULES := proxy logger
BIN := cmd/server cmd/local

GITTAG := `git describe --tags`
VERSION := `git describe --abbrev=0 --tags`
RELEASE := `git rev-list $(shell git describe --abbrev=0 --tags).. --count`
BUILD_TIME := `date +%FT%T%z`
# Setup the -ldflags option for go build here, interpolate the variable values
LDFLAGS := -ldflags "-X main.GitTag=${GITTAG} -X main.BuildTime=${BUILD_TIME}"

vendor:
	go get ./...
	go get github.com/stretchr/testify/assert


test:
	go test ./... --race -cover;

fmt:
	find . -name "*.go" -type f -exec echo {} \;  |\
	while IFS= read -r line; \
	do \
		echo "$$line";\
		goimports -w "$$line" "$$line";\
	done

build:
	mkdir -p bin;\
	echo ==================================; \
	for m in $(BIN); do \
		cd $(PWD)/$$m && go build ${LDFLAGS} -o ../../bin/$$m --race ; \
	done
	echo ==================================; \
	cd $(PWD) && cp gen_key_cert.sh ./bin

install: vendor build

deploy:
	for m in $(BIN); do \
		cd $(PWD)/$$m && gox ${LDFLAGS} -osarch="linux/amd64" -output ../../dist/{{.OS}}_{{.Arch}}_{{.Dir}};\
		cd $(PWD)/$$m && gox ${LDFLAGS} -os="windows" -output ../../dist/{{.OS}}_{{.Arch}}_{{.Dir}};\
		cd $(PWD)/$$m && gox ${LDFLAGS} -osarch="darwin/amd64" -output ../../dist/{{.OS}}_{{.Arch}}_{{.Dir}};\
	done
