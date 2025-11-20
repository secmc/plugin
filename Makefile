.PHONY: run proto

run:
	cd cmd && go run .
proto:
	cd proto && buf generate
