name = setlxPlayground

package = github.com/KeKsBoTer/$(name)/app

all: build run

run:
	./$(name).exe -mode="dev"

build:
	go build -o $(name).exe $(package)

deps:
	go get github.com/gorilla/mux