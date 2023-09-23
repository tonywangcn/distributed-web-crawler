# Dockerfile
FROM golang:1.20-buster AS builder
RUN mkdir /src
ADD ./go/src/ /src
WORKDIR /src

RUN go env -w GO111MODULE=auto
RUN go build -o main .

FROM gcr.io/distroless/base-debian10

WORKDIR /

COPY --from=builder /src/main /main
EXPOSE 3000 8000