GOFMT_FILES?=$$(find . -name '*.go' |grep -v vendor)

MAKEFLAGS += --silent

build:
	go build -o terraform-provider-keycloak
	# go build -gcflags="all=-N -l" -o terraform-provider-keycloak

build-example-arm: build
	mkdir -p example/.terraform/plugins/terraform.local/mrparkers/keycloak/3.0.0/darwin_arm64
	mkdir -p example/terraform.d/plugins/terraform.local/mrparkers/keycloak/3.0.0/darwin_arm64
	cp terraform-provider-keycloak example/.terraform/plugins/terraform.local/mrparkers/keycloak/3.0.0/darwin_arm64/
	cp terraform-provider-keycloak example/terraform.d/plugins/terraform.local/mrparkers/keycloak/3.0.0/darwin_arm64/

build-example: build
	mkdir -p example/.terraform/plugins/terraform.local/mrparkers/keycloak/3.0.0/darwin_amd64
	mkdir -p example/terraform.d/plugins/terraform.local/mrparkers/keycloak/3.0.0/darwin_amd64
	cp terraform-provider-keycloak example/.terraform/plugins/terraform.local/mrparkers/keycloak/3.0.0/darwin_amd64/
	cp terraform-provider-keycloak example/terraform.d/plugins/terraform.local/mrparkers/keycloak/3.0.0/darwin_amd64/

local: deps
	docker-compose up --build -d
	./scripts/wait-for-local-keycloak.sh
	./scripts/create-terraform-client.sh

deps:
	./scripts/check-deps.sh

fmt:
	gofmt -w -s $(GOFMT_FILES)

test: fmtcheck vet
	go test $(TEST)

testacc: fmtcheck vet
	go test -v github.com/mrparkers/terraform-provider-keycloak/keycloak
	TF_ACC=1 CHECKPOINT_DISABLE=1 go test -v -timeout 60m -parallel 4 github.com/mrparkers/terraform-provider-keycloak/provider $(TESTARGS)

testacc-up: fmtcheck vet
	TF_ACC=1 CHECKPOINT_DISABLE=1 KEYCLOAK_CLIENT_ID=terraform KEYCLOAK_CLIENT_SECRET=884e0f95-0f42-4a63-9b1f-94274655669e KEYCLOAK_CLIENT_TIMEOUT=5 KEYCLOAK_REALM=master KEYCLOAK_URL="http://localhost:8080" go test -timeout 60m -run ^TestAccKeycloakRealmUserProfile_basicFull github.com/mrparkers/terraform-provider-keycloak/provider $(TESTARGS)

fmtcheck:
	lineCount=$(shell gofmt -l -s $(GOFMT_FILES) | wc -l | tr -d ' ') && exit $$lineCount

vet:
	go vet ./...

user-federation-example:
	cd custom-user-federation-example && ./gradlew shadowJar
