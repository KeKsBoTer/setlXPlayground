name = setlxPlayground

package = ./app

all: build run

run:
	./$(name).exe -mode="dev"

build:
	go build -o $(name).exe $(package)

deps:
	go get github.com/gorilla/mux

release:
	CGO_ENABLED=0 go build -ldflags="-s -w" -a -installsuffix nocgo -o setlxplay ./app