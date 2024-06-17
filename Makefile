run:
	@echo Make commands:
	@echo test - run full tests [unit+intergration] and create coverage files
	@echo test-unit - run only unit-tests [no intergration] and create coverage files
	@echo test-verbose - run full tests with -v console output
	@echo gen - [re]generage swagger documentation and mocks
	@echo stop - stopping docker containers
	@echo start - full application starting [run tests and docker]
	@echo deps - install project depedencies
	@echo down - down the docker containers
	@echo upgrade - upgrade and download all depedencies, then run tests [if tests are failed, changes will be canceled]
	@echo ! When running make test make sure docker is up
	@echo Example starting usage: make gen start

test-verbose:
	@cd ./api_gateway/ && go generate ./... && go test ./... -v
	@cd ./api_user/ && go generate ./... && go test ./... -v
	@cd ./api_notification/ && go generate ./... && go test ./... -v

test: test-folder-creation
	@cd ./api_gateway/ && go generate ./... && go test ./... -coverprofile=tests/coverage -coverpkg=./... -json | go-test-report -o tests/report.html -t Gateway-Testing-Results && go tool cover -func=tests/coverage -o tests/coverage.func && go tool cover -html=tests/coverage -o tests/coverage.html
	@cd ./api_user/ && go generate ./... && go test ./... -coverprofile=tests/coverage -coverpkg=./... -json | go-test-report -o tests/report.html -t User-Testing-Results && go tool cover -func=tests/coverage -o tests/coverage.func && go tool cover -html=tests/coverage -o tests/coverage.html
	@cd ./api_notification/ && go generate ./... && go test ./... -coverprofile=tests/coverage -coverpkg=./... -json | go-test-report -o tests/report.html -t Notification-Testing-Results && go tool cover -func=tests/coverage -o tests/coverage.func && go tool cover -html=tests/coverage -o tests/coverage.html

test-unit: test-folder-creation
	@cd ./api_gateway/ && go generate ./... && go test ./... -short -coverprofile=tests/coverage -coverpkg=./... -json | go-test-report -o tests/report.html -t Gateway-Testing-Results && go tool cover -func=tests/coverage -o tests/coverage.func && go tool cover -html=tests/coverage -o tests/coverage.html
	@cd ./api_user/ && go generate ./... && go test ./... -short -coverprofile=tests/coverage -coverpkg=./... -json | go-test-report -o tests/report.html -t User-Testing-Results && go tool cover -func=tests/coverage -o tests/coverage.func && go tool cover -html=tests/coverage -o tests/coverage.html
	@cd ./api_notification/ && go generate ./... && go test ./... -short -coverprofile=tests/coverage -coverpkg=./... -json | go-test-report -o tests/report.html -t Notification-Testing-Results && go tool cover -func=tests/coverage -o tests/coverage.func && go tool cover -html=tests/coverage -o tests/coverage.html

test-folder-creation:
ifeq ($(OS),Windows_NT)
	@cd ./api_gateway/ && mkdir tests & echo.
	@cd ./api_user/ && mkdir tests & echo.
	@cd ./api_notification/ && mkdir tests & echo.
else
	@cd ./api_gateway/ && mkdir -p tests
	@cd ./api_user/ && mkdir -p tests
	@cd ./api_notification/ && mkdir -p tests
endif

gen:
	@swag init --parseDependency -d ./api_gateway/internal/handlers -g ../../cmd/gateway/main.go -o ./api_gateway/docs
	@cd ./api_gateway/ && go generate ./...
	@cd ./api_user/ && go generate ./...
	@cd ./api_notification/ && go generate ./...

stop:
	@docker compose stop

down:
	@docker compose down

start:
	@make test-verbose
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
	@$(MAKE) deps

	@cd ./api_gateway/ && go get -u ./... && go mod tidy
	@cd ./api_user/ && go get -u ./... && go mod tidy
	@cd ./api_notification/ && go get -u ./... && go mod tidy
	
	-@$(MAKE) test-verbose
ifeq ($(OS),Windows_NT)
	ifneq %errorlevel% 0
	    echo "*** Tests are failed, canceling the upgrade ***"
		git reset
		git checkout .
		false
	endif
else
	@if [ $$? -ne 0 ]; \
    then \
        echo "*** Tests are failed, canceling the upgrade ***"; \
		git reset; \
		git checkout .; \
        false; \
    fi
endif
	@echo All depedencies upgraded successfully;