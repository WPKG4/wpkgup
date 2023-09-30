FROM golang:latest AS build

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o wpkgup .

FROM alpine:latest

WORKDIR /app

COPY --from=build /app/wpkgup .

EXPOSE 8080

VOLUME /app/data
CMD ["./wpkgup","server","-w","/app/data"]
