FROM golang:1.12 as builder
WORKDIR /server/

COPY go.mod go.mod
COPY go.sum go.sum

RUN go mod download

COPY  cmd cmd
COPY  execute.go .
COPY  handler.go .
COPY  router.go .
COPY  database.go .

RUN CGO_ENABLED=0 go build -ldflags="-s -w" -a -installsuffix nocgo -o setlxplay github.com/keksboter/setlxplayground/cmd/setlxplayground

FROM gcr.io/distroless/java
WORKDIR /root/
COPY setlx setlx
COPY www www
COPY java.policy java.policy
COPY --from=builder /server/setlxplay .
ENTRYPOINT [ "./setlxplay","-mode","prod","-database","db"]
EXPOSE 80