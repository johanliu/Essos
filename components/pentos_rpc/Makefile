all: main

main:
	@go build --buildmode=plugin -o /opt/essos/pipeline.so pentos_rpc.go

clean:
	@rm -rf dist

.PHONY: all main clean debug
