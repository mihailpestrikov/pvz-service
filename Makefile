build:
	docker-compose --env-file .env -f docker-compose.yml build

up: build
	docker-compose --env-file .env -f docker-compose.yml up -d

test:
	docker-compose -f docker-compose.test.yml up --build --abort-on-container-exit integration-tests

unit-test:
	go test -v ./internal/... ./pkg/...

down:
	docker-compose -f docker-compose.yml down
