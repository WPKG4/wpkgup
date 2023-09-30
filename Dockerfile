FROM golang:latest AS build

WORKDIR /app

COPY . .

RUN make build

FROM debian:sid-slim

WORKDIR /app

COPY --from=build /app/wpkgup /app

EXPOSE 8080

VOLUME /app/data
ENTRYPOINT ["./wpkgup","server","-w","/app/data"]
