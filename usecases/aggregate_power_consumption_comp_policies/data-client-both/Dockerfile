FROM golang:1.11.4 AS builder

# Download and install the latest release of dep
ADD https://github.com/golang/dep/releases/download/v0.4.1/dep-linux-amd64 /usr/bin/dep
RUN chmod +x /usr/bin/dep

# Copy the code from the host and compile it
WORKDIR $GOPATH/src/github.com/JacobMoxham/PartIIProjectImplementation/usecases/aggregate_power_consumption/data-client
COPY Gopkg.toml Gopkg.lock ./
RUN dep ensure --vendor-only
COPY . ./
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix nocgo -o /app .

FROM scratch
COPY --from=builder /app ./
ENTRYPOINT ["./app"]

## Debug Version
#FROM golang:1.11.4-alpine3.7 AS builder
#
## Download and install the latest release of dep
#ADD https://github.com/golang/dep/releases/download/v0.4.1/dep-linux-amd64 /usr/bin/dep
#RUN chmod +x /usr/bin/dep
#
## Copy the code from the host and compile it
#WORKDIR $GOPATH/src/github.com/JacobMoxham/PartIIProjectImplementation/usecases/aggregate_power_consumption_data_policies/data-client-raw-data
#COPY Gopkg.toml Gopkg.lock ./
#RUN dep ensure --vendor-only
#COPY . ./
#RUN CGO_ENABLED=0 GOOS=linux go build -gcflags "all=-N -l" -a -installsuffix nocgo -o /app .
#
## Compile Delve
#RUN apk add --no-cache git
#RUN go get github.com/derekparker/delve/cmd/dlv
#
## Final stage
#FROM alpine:3.7
#
#EXPOSE 40000
#
## Allow delve to run on Alpine based containers.
#RUN apk add --no-cache libc6-compat
#
#FROM scratch
#COPY --from=builder /app ./
#COPY --from=builder /go/bin/dlv /
#
## Run delve
#CMD ["/dlv", "--listen=:40000", "--headless=true", "--api-version=2", "exec", "/app"]