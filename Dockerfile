FROM golang:1.17

COPY . /raptor

WORKDIR /raptor

RUN go build

ENTRYPOINT ["./raptor"]