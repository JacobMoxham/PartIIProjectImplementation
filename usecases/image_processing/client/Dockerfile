#FROM golang:1.11.4 AS builder
#
## Download and install the latest release of dep
#ADD https://github.com/golang/dep/releases/download/v0.4.1/dep-linux-amd64 /usr/bin/dep
#RUN chmod +x /usr/bin/dep
#
## Copy the code from the host and compile it
#WORKDIR $GOPATH/src/github.com/JacobMoxham/PartIIProjectImplementation/usecases/image_processing/client
#COPY Gopkg.toml Gopkg.lock ./
#RUN dep ensure --vendor-only
#COPY . ./
#RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix nocgo -o /app .
#
#FROM scratch
#COPY --from=builder /app ./
#ENTRYPOINT ["./app"]

# Back to basics
FROM gildasch/tensorflow-go as builder

RUN mkdir -p /model && \
  curl -o /model/inception5h.zip -s "http://download.tensorflow.org/models/inception5h.zip" && \
  unzip /model/inception5h.zip -d /model

# Download and install the latest release of dep
ADD https://github.com/golang/dep/releases/download/v0.4.1/dep-linux-amd64 /usr/bin/dep
RUN chmod +x /usr/bin/dep

# Copy the code from the host and compile it
WORKDIR $GOPATH/src/github.com/JacobMoxham/PartIIProjectImplementation/usecases/image_processing/client

COPY Gopkg.toml Gopkg.lock ./
RUN dep ensure --vendor-only
COPY . .
RUN go build
ENTRYPOINT ["./client"]