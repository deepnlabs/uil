.PHONY: all build plugin clean

all: build plugin

build:
	go build -o bin/uild cmd/uild/main.go

plugin:
	go build -buildmode=plugin -o plugins/custom_interlock.so plugins_src/custom_interlock/main.go

clean:
	rm -rf bin/ plugins/custom_interlock.so
