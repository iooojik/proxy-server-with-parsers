FROM golang:1.18 as build
COPY . /var/www
WORKDIR /var/www/
RUN go mod download
RUN go test ./...
RUN go build main.go