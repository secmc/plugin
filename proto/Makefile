.PHONY: proto proto-go proto-php proto-ts lint breaking clean

proto: proto-go proto-php proto-ts

proto-go:
	buf generate && cd .. && go mod tidy

proto-php:
	buf generate --template buf.gen.php.yaml

proto-ts:
	buf generate --template buf.gen.ts.yaml

proto-py:
	buf generate --template buf.gen.py.yaml

lint:
	buf lint

breaking:
	buf breaking --against '.git#branch=main'

clean:
	rm -rf generated/*.pb.go generated/php/ generated/ts/