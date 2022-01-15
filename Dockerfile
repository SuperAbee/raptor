FROM golang:1.15

COPY . /raptor

WORKDIR /raptor

RUN go build

ENTRYPOINT ["./raptor"]