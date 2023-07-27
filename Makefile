BIN_DIR = bin
GOMODFILE ?= go.mod
PROTO_DIR = proto
FRONTEND_DIR = web/src/components/SelfHosted/proto
PACKAGE = $(cd proto/ && shell head -1 go.mod | awk '{print $$2}')
GO = $(HOME)/go/bin/go1.18.10

gen.proto:
	protoc -I${PROTO_DIR} --go_opt=module=${PACKAGE} --go_out=. --go-grpc_opt=module=${PACKAGE} --go-grpc_out=. ${PROTO_DIR}/*.proto
	protoc -I${PROTO_DIR} --grpc-gateway_out ${PROTO_DIR} \
        --grpc-gateway_opt logtostderr=true \
        --grpc-gateway_opt paths=source_relative \
        ${PROTO_DIR}/*.proto
	protoc \
		--grpc-gateway-ts_out=loglevel=debug,use_proto_names=true:${FRONTEND_DIR} \
		--proto_path=${PROTO_DIR} ${PROTO_DIR}/query_explainer.proto

gen.types:
	cd postgres-explain-core && make generate-types

gen.wasm: gen.types
	(cd wasm && go mod tidy && GOOS=js GOARCH=wasm go build -o ../public/main.wasm main.go)

build.backend:
	(cd bm-server/ && go mod tidy -modfile=$(GOMODFILE) && GOOS=linux GOARCH=amd64 go build -modfile=$(GOMODFILE) -o bin/bm-server .)