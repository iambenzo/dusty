# Builder image
FROM golang:alpine AS builder
# Git is required for fetching the dependencies.
RUN apk update && apk add --no-cache git ca-certificates && update-ca-certificates
WORKDIR $GOPATH/src/iambenzo.com/dusty/
COPY . .
RUN go get -d -v
# RUN go build -o /go/bin/dusty
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /go/bin/dusty

# Smallish image for actual running
FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/bin/dusty /go/bin/dusty
ENTRYPOINT ["/go/bin/dusty"]
