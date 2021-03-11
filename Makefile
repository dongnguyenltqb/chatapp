default: kill build run

dev:
	@nodemon -e go --exec make

kill:
	@pkill "gapp" || true

build:
	@go build
run:
	@./gapp
