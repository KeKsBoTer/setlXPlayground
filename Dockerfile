FROM golang:1.11 as builder
WORKDIR /server/

COPY go.mod go.mod
COPY go.sum go.sum

RUN go mod download

COPY  main.go main.go
COPY  run.go run.go
COPY  database.go database.go

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -a -installsuffix nocgo -o setlxplay .


FROM gcr.io/distroless/java
WORKDIR /root/
COPY setlx setlx
COPY www www
COPY java.policy java.policy
COPY --from=builder /server/setlxplay .
ENTRYPOINT [ "./setlxplay","-mode","prod","-database","db"]
EXPOSE 8080
