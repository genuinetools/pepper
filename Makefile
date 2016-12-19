OS = darwin freebsd linux openbsd windows
ARCHS = 386 arm amd64 arm64

all: build release

build: deps
	go build

release: clean deps
	@for arch in $(ARCHS);\
	do \
		for os in $(OS);\
		do \
			echo "Building $$os-$$arch"; \
			mkdir -p build/pepper-$$os-$$arch/; \
			GOOS=$$os GOARCH=$$arch go build -o build/pepper-$$os-$$arch/pepper; \
			tar cz -C build -f build/pepper-$$os-$$arch.tar.gz pepper-$$os-$$arch; \
		done \
	done

test: deps
	go test ./...

deps:
	go get -d -v -t ./...

clean:
	rm -rf build
	rm -f pepper
