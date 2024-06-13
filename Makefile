run:
	@echo Make commands:
	@echo test - start project testing and create coverage files
	@echo test-verbose - testing with -v flag, but with no coverage
	@echo gen - (re)generage swagger documentation for gateway
	@echo stop - stopping docker containers
	@echo start - full application starting (run tests and docker)
	@echo.
	@echo Example starting usage: make gen start

test:
	@cd ./api_gateway/ && go test ./... -coverprofile=coverage -coverpkg=./... && go tool cover -func=coverage -o coverage.func && go tool cover -html=coverage -o coverage.html
	@cd ./api_user/ && go test ./... -coverprofile=coverage -coverpkg=./... && go tool cover -func=coverage -o coverage.func && go tool cover -html=coverage -o coverage.html
	@cd ./api_notification/ && go test ./... -coverprofile=coverage -coverpkg=./... && go tool cover -func=coverage -o coverage.func && go tool cover -html=coverage -o coverage.html

test-verbose:
	@cd ./api_gateway/ && go test ./... -v
	@cd ./api_user/ && go test ./... -v
	@cd ./api_notification/ && go test ./... -v

gen:
	@swag init --parseDependency -d ./api_gateway/internal/handlers -g ../../cmd/gateway/main.go -o ./api_gateway/docs

stop:
	@docker compose stop

start:
	@make test
	@docker compose up --build --timestamps --wait --wait-timeout 1800 --remove-orphans -d