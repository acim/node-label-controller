FROM golang:1.13-alpine AS builder

ARG REVISION

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -a -installsuffix cgo -ldflags="-s -w" -o /go/bin/node-label-controller

FROM alpine

ARG CREATED
ARG REVISION
ARG VERSION

RUN adduser -D ablab

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /go/bin/node-label-controller /usr/bin/node-label-controller

LABEL org.opencontainers.image.authors="boban.acimovic@gmail.com" \
    org.opencontainers.image.created=$CREATED \
    org.opencontainers.image.description="node-label-controller image based on alpine" \
    org.opencontainers.image.documentation="https://github.com/acim/node-label-controller/blob/master/README.md" \
    org.opencontainers.image.revision=$REVISION \
    org.opencontainers.image.source="https://github.com/acim/node-label-controller" \
    org.opencontainers.image.title="node-label-controller" \
    org.opencontainers.image.vendor="ablab.de" \
    org.opencontainers.image.version=$VERSION

USER ablab

ENTRYPOINT ["/usr/bin/node-label-controller"]