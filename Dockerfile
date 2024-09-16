#### BUILD STAGE
FROM golang:1.22-alpine AS builder
RUN apk add --update --no-cache build-base

WORKDIR $GOPATH/src/back/

# Copy project files into the builder
COPY . .

RUN go build -a -mod=vendor -ldflags '-linkmode external -w -s -extldflags "-static"' -o ./gateway

#### IMAGE DEFINITION
FROM alpine:3

# Upgrade system dependencies for security patches or bug fixing
RUN apk update && apk upgrade
WORKDIR /app
# Copy the programm with for uid 1000 and gid 1000
COPY --chown=1000:1000 --from=builder /go/src/back/gateway gateway


# Run with a non-root user
USER 1000:1000

CMD ["./gateway"]