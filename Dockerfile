FROM golang:1.16-alpine as builder

WORKDIR /build

RUN apk add --no-cache make git upx

COPY . .

RUN make install-tools \
  && make build \
  && upx --best --lzma bin/spark-nanny

FROM scratch

COPY --from=builder /build/bin/spark-nanny /spark-nanny

ENTRYPOINT ["/spark-nanny"]
