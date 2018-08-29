FROM golang as builder
WORKDIR /server/
COPY app .

RUN go get github.com/gorilla/mux
RUN go build -a -o setlxplay .


# FROM gcr.io/distroless/java
FROM anapsix/alpine-java
WORKDIR /root/
COPY setlx setlx
COPY www www
COPY --from=builder /server/setlxplay .
ENV PATH "$PATH:/root/setlx"
ENTRYPOINT [ "./setlxplay" ]
EXPOSE 8080