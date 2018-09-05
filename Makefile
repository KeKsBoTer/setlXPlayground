name = setlxPlayground

package = .

all: build run

run:
	./$(name).exe -port="8080" -mode="dev"

build:
	GO111MODULE=on go build -o $(name).exe $(package)

release:
	CGO_ENABLED=0 GO111MODULE=on go build -ldflags="-s -w" -a -installsuffix nocgo -o setlxplay $(package)