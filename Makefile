.PHONY: core
BIN_DIR = bin
GOMODFILE ?= go.mod
PROTO_DIR = proto
WEB_DIR = web
FRONTEND_DIR = web/src/components/SelfHosted/proto
PACKAGE = $(cd proto/ && shell head -1 go.mod | awk '{print $$2}')
GO = $(HOME)/go/bin/go1.18.10

submodules:
	 git submodule update --recursive --remote

gen.proto:
	protoc -I${PROTO_DIR} --go_opt=module=${PACKAGE} --go_out=. --go-grpc_opt=module=${PACKAGE} --go-grpc_out=. ${PROTO_DIR}/*.proto
	protoc -I${PROTO_DIR} --grpc-gateway_out ${PROTO_DIR} \
        --grpc-gateway_opt logtostderr=true \
        --grpc-gateway_opt paths=source_relative \
        ${PROTO_DIR}/*.proto
	protoc \
		--grpc-gateway-ts_out=loglevel=debug,use_proto_names=true:${FRONTEND_DIR} \
		--proto_path=${PROTO_DIR} ${PROTO_DIR}/query_explainer.proto ${PROTO_DIR}/info.proto ${PROTO_DIR}/analytics.proto

gen.types:
	cd core && make generate-types

gen.wasm: gen.types
	(cd wasm && go mod tidy && GOOS=js GOARCH=wasm go build -o ../${WEB_DIR}/public/main.wasm main.go)

gen.core:
	./core.sh

gen.sanitize:
	curl -o backend/shared/sanitize.go https://raw.githubusercontent.com/jackc/pgx/master/internal/sanitize/sanitize.go
	sed -i 's/package sanitize/package shared/' backend/shared/sanitize.go

build.backend:
	(cd backend/ && go mod tidy -modfile=$(GOMODFILE) && GOOS=linux GOARCH=amd64 go build -modfile=$(GOMODFILE) -o bin/backend .)

build.collector:
	(cd collector/ && go mod tidy -modfile=$(GOMODFILE) && GOOS=linux GOARCH=amd64 go build -modfile=$(GOMODFILE) -o bin/collector .)

build.backend.core: gen.core build.backend

run: build.backend.core
	docker-compose up --build --remove-orphans

run.frontend:
	(cd web && REACT_APP_MODE=self_hosted REACT_APP_BACKEND_ORIGIN=http://localhost:8082 npm run start)

run.web:
	(cd web && npm run start)

reload.backend: build.backend.core
	docker-compose up --build --remove-orphans backend -d