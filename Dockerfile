FROM golang:1.11 as builder
WORKDIR /server/

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -a -installsuffix nocgo -o setlxplay .


# FROM gcr.io/distroless/java
FROM anapsix/alpine-java
WORKDIR /root/
COPY setlx setlx
COPY www www
COPY --from=builder /server/setlxplay .
ENV PATH "$PATH:/root/setlx"
ENTRYPOINT [ "./setlxplay","-mode","prod"]
EXPOSE 8080
