FROM golang:1-alpine AS build-env

RUN apk --no-cache add gcc

WORKDIR ${GOPATH}/src/github.com/0xERR0R/fritzbox-rdns

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /go/bin/fritzbox-rdns .

FROM alpine

LABEL org.opencontainers.image.source="https://github.com/0xERR0R/fritzbox-rdns" \
      org.opencontainers.image.url="https://github.com/0xERR0R/fritzbox-rdns" \
      org.opencontainers.image.title="rDNS server for FritzBox"

COPY --from=build-env /go/bin/fritzbox-rdns /app/fritzbox-rdns

EXPOSE 53

ENTRYPOINT ["/app/fritzbox-rdns"]