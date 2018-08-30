name = setlxPlayground

package = ./app

all: build run

run:
	./$(name).exe

build:
	go build -o $(name).exe $(package)