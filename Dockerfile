FROM golang:alpine
RUN mkdir /app
ADD main.go /app/
ADD config/config.ini /app/config/
WORKDIR /app
RUN apk update; apk add git; go get -d ./... ; go build -o main .
CMD ["./main"]
