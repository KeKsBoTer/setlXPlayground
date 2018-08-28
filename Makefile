name = setlxPlayground

package = github.com/KeKsBoTer/$(name)/app

all: build run

run:
	./$(name).exe

build:
	go build -o $(name).exe $(package)