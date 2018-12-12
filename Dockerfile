# Building reMP3
FROM alpine:edge AS go
WORKDIR /go/src/reMP3
RUN echo "http://dl-cdn.alpinelinux.org/alpine/edge/testing" >> /etc/apk/repositories
RUN apk add --no-cache go git fftw-dev musl-dev dep
ENV GOPATH /go
COPY Gopkg.* ./
COPY *.go ./
RUN dep ensure
RUN go build -o reMP3 *.go


# Create Release image without dev dependencies
FROM alpine:edge AS release
WORKDIR /usr/local/bin/
RUN echo "http://dl-cdn.alpinelinux.org/alpine/edge/testing" >> /etc/apk/repositories
RUN apk add --no-cache ca-certificates ffmpeg
COPY --from=go /go/src/reMP3/reMP3 .
ENV CFG_LISTEN ":7090"
CMD ["./reMP3"]
