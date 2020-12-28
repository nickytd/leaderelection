### build go executable
FROM golang:1.15

COPY . /go/src/leaderelection
WORKDIR /go/src/leaderelection
RUN GO111MODULE=on CGO_ENABLED=0 GOOS=linux go build -o leaderelection

### final image
FROM alpine:latest

WORKDIR /

RUN apk add --no-cache tini bash
COPY --from=0 /go/src/leaderelection/leaderelection /leaderelection
RUN chmod 755 /leaderelection

ENV clusters ""
ENTRYPOINT ["/sbin/tini", "--", "/leaderelection"]
