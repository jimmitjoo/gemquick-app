BINARY_NAME=gemquickApp

build:
	@go mod vendor
	@echo "Building Gemquick..."
	@go build -o tmp/${BINARY_NAME} .
	@echo "Gemquick built!"

run: build
	@echo "Starting Gemquick..."
	@./tmp/${BINARY_NAME} &
	@echo "Gemquick started!"

clean:
	@echo "Cleaning..."
	@go clean
	@rm tmp/${BINARY_NAME}
	@echo "Cleaned!"

start_compose:
	docker-compose up -d

stop_compose:
	docker-compose down

test:
	@echo "Testing..."
	@go test ./...
	@echo "Done!"

start: run

stop:
	@echo "Stopping Gemquick..."
	@-pkill -SIGTERM -f "./tmp/${BINARY_NAME}"
	@echo "Stopped Gemquick!"

restart: stop start