FROM golang as builder
WORKDIR /server/

COPY go.mod .
COPY go.sum .
COPY app .

RUN go build -a -o setlxplay ./app


# FROM gcr.io/distroless/java
FROM anapsix/alpine-java
WORKDIR /root/
COPY setlx setlx
COPY www www
COPY --from=builder /server/setlxplay .
ENV PATH "$PATH:/root/setlx"
ENTRYPOINT [ "./setlxplay" ]
EXPOSE 8080