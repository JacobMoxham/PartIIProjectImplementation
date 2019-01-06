#FROM golang:alpine AS builder

## This works by pulling from github
#RUN apk add --no-cache git
#RUN go get github.com/JacobMoxham/PartIIProjectImplementation
#WORKDIR src/github.com/JacobMoxham/PartIIProjectImplementation
#RUN go build -o power-example .
#EXPOSE 3001
#CMD ["./power-example"]

#FROM golang:alpine AS builder

## This (doesn't) works by building locally
#RUN mkdir /build
#RUN echo $(ls src)
#ADD . /build/
#WORKDIR /build
#RUN go build -o main .
#
#FROM alpine
##RUN adduser -S -D -H -h /app appuser
##USER appuser
#COPY --from=builder /build/main /app/
#WORKDIR /app
#CMD ["./main"]

FROM golang:1.11 AS builder

# Download and install the latest release of dep
ADD https://github.com/golang/dep/releases/download/v0.4.1/dep-linux-amd64 /usr/bin/dep
RUN chmod +x /usr/bin/dep

# Copy the code from the host and compile it
WORKDIR $GOPATH/src/github.com/JacobMoxham/PartIIProjectImplementation
COPY Gopkg.toml Gopkg.lock ./
RUN dep ensure --vendor-only
COPY . ./
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix nocgo -o /app .

FROM scratch
COPY --from=builder /app ./
ENTRYPOINT ["./app"]
