FROM golang:alpine AS build

# install tools required
RUN apk add --no-cache git
RUN go get github.com/JacobMoxham/PartIIProjectImplementation
WORKDIR /github.com/JacobMoxham/PartIIProjectImplementation/usecases

RUN go build -o power-example .
CMD ["./power-example"]

