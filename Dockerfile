# Build vsphere-monitor in a separate container
FROM golang:1.11 AS builder

RUN go get github.com/FiloSottile/gvt
RUN update-ca-certificates

WORKDIR /go/src/github.com/travis-ci/vsphere-monitor

COPY . .
RUN make deps
ENV CGO_ENABLED 0
RUN make build

FROM scratch

# Copy things from the other stages
COPY --from=builder /go/bin/vsphere-monitor .
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

ENTRYPOINT ["./vsphere-monitor"]
