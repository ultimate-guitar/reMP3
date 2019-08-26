# Building reMP3
FROM alpine:edge AS go
WORKDIR /reMP3
RUN echo "http://dl-cdn.alpinelinux.org/alpine/edge/testing" >> /etc/apk/repositories
RUN apk add --no-cache go git fftw-dev musl-dev
COPY . ./
RUN go mod vendor
RUN go build -o reMP3 *.go

# Create Release image without dev dependencies
FROM alpine:edge AS release
WORKDIR /usr/local/bin/
RUN echo "http://dl-cdn.alpinelinux.org/alpine/edge/testing" >> /etc/apk/repositories
RUN apk add --no-cache ca-certificates ffmpeg
COPY --from=go /reMP3/reMP3 .
ENV CFG_LISTEN ":8080"
CMD ["./reMP3"]
