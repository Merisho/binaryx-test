FROM golang:1.16
WORKDIR /go/src/github.com/merisho/binaryx-test
COPY . ./
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o app main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=0 /go/src/github.com/merisho/binaryx-test/app ./
CMD ["./app"]
