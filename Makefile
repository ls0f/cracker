

MODULES := proxy server local
BIN := server local

vendor:
	for m in $(MODULES) ; do \
	cd $$m && go get -insecure -v && cd -;\
	done


test:
	echo ==================================; \
	for m in $(MODULES); do \
		cd $(PWD)/$$m && go test --race -cover; \
		echo ==================================; \
	done

fmt:
	find . -name "*.go" -type f -exec echo {} \; | grep -v -E "github.com|gopkg.in"|\
	while IFS= read -r line; \
	do \
		echo "$$line";\
		goimports -w "$$line" "$$line";\
	done

build:
	mkdir -p bin;\
	echo ==================================; \
	for m in $(BIN); do \
		cd $(PWD)/$$m && go build -o ../bin/$$m --race ; \
	done
	echo ==================================; \

