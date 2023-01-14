FROM golang:1.19 as build-stage
RUN apt-get update
WORKDIR /build
COPY ./ /build/
RUN go get -u github.com/golang/protobuf/protoc-gen-go && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
ENV PATH=$PATH:$GOPATH/bin:/opt/protoc/bin
RUN apt-get update && \
  apt-get install -y \
  git \
  protobuf-compiler \
  golang-goprotobuf-dev \
  ca-certificates && \
  apt-get autoremove -y && \
  apt-get clean -y && \
  rm -rf /var/cache/apt/archives /var/lib/apt/lists/*
RUN go mod download
RUN ./generate_stubs.sh
RUN go build ./cmd/main.go
FROM ubuntu:22.04 as final-stage
WORKDIR /app
RUN apt-get update && \
  apt-get install -y \
  ca-certificates && \
  apt-get autoremove -y && \
  apt-get clean -y && \
  rm -rf /var/cache/apt/archives /var/lib/apt/lists/*
COPY --from=build-stage ./build/ .
CMD ["./main"] 
