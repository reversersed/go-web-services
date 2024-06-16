run:
	@echo Make commands:
	@echo test - start project testing and create coverage files
	@echo test-verbose - testing with -v flag, but with no coverage
	@echo gen - [re]generage swagger documentation for gateway
	@echo stop - stopping docker containers
	@echo start - full application starting [run tests and docker]
	@echo deps - install project depedencies
	@echo down - down the docker containers
	@echo upgrade - upgrade and download all depedencies, then run tests [if tests are failed, changes will be canceled]
	@echo Example starting usage: make gen start

test:
	@cd ./api_gateway/ && go generate ./... && go test ./... -coverprofile=coverage -coverpkg=./... && go tool cover -func=coverage -o coverage.func && go tool cover -html=coverage -o coverage.html
	@cd ./api_user/ && go generate ./... && go test ./... -coverprofile=coverage -coverpkg=./... && go tool cover -func=coverage -o coverage.func && go tool cover -html=coverage -o coverage.html
	@cd ./api_notification/ && go generate ./... && go test ./... -coverprofile=coverage -coverpkg=./... && go tool cover -func=coverage -o coverage.func && go tool cover -html=coverage -o coverage.html

test-verbose:
	@cd ./api_gateway/ && go test ./... -v
	@cd ./api_user/ && go test ./... -v
	@cd ./api_notification/ && go test ./... -v

gen:
	@swag init --parseDependency -d ./api_gateway/internal/handlers -g ../../cmd/gateway/main.go -o ./api_gateway/docs

stop:
	@docker compose stop

down:
	@docker compose down

start:
	@make test
	@docker compose up --build --timestamps --wait --wait-timeout 1800 --remove-orphans -d

deps:
	@go install github.com/golang/mock/mockgen@latest
	@go install github.com/swaggo/swag/cmd/swag@latest

	@cd ./api_gateway/ && go mod tidy
	@cd ./api_notification/ && go mod tidy
	@cd ./api_user/ && go mod tidy
ifneq ($(OS), Windows_NT)
	@export PATH=$(PATH):$(HOME)/go/bin
endif
	@echo all depedencies are installed

upgrade:
	@cd ./api_gateway/ && go get -u ./... && go mod tidy
	@cd ./api_user/ && go get -u ./... && go mod tidy
	@cd ./api_notification/ && go get -u ./... && go mod tidy
	
	@make test; \
	if [ $$? -ne 0 ]; \
    then \
        echo "*** Tests are failed, canceling the upgrade ***"; \
		git reset; \
		git checkout .; \
        false; \
    fi
	@echo All depedencies upgraded successfully;