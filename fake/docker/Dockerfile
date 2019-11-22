# build the server binary
FROM golang:1.13.3 AS builder
LABEL stage=server-intermediate
WORKDIR /go/src/github.com/infobloxopen/themis/
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o ./fake/bin/server ./fake/cmd/server/


# copy the server binary from builder stage; run the server binary
FROM alpine:latest AS runner
WORKDIR /bin
COPY --from=builder /go/src/github.com/infobloxopen/themis/fake/bin/server .
COPY ./fake/docker/entrypoint.sh /bin


ENTRYPOINT ["sh", "entrypoint.sh"]
