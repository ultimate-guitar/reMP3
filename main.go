package main

import (
	"github.com/buaazp/fasthttprouter"
	"github.com/namsral/flag"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/reuseport"
	"log"
	"time"
)

type Config struct {
	WebListen            string
	FFmpegBinary         string
	ServerConvertTimeout time.Duration
}

const (
	resizeHeaderNameSource        = "x-resize-base"
	resizeHeaderNameSchema        = "x-resize-scheme"
	resizeHeaderDefaultSchema     = "https"
	resizeQArgBitrateName         = "bitrate"
	resizeQArgOutSampleRateName   = "samplerate"
	resizeQArgDurationName        = "duration"
	httpClientMaxIdleConns        = 64
	httpClientMaxIdleConnsPerHost = 64
	httpClientIdleConnTimeout     = 120 * time.Second
	httpClientFileDownloadTimeout = 30 * time.Second
	serverMaxConcurrencyRequests  = 64
	serverRequestReadTimeout      = 10 * time.Second
	serverResponseWriteTimeout    = 20 * time.Second
	serverMaxBodySize             = 100 * 1024 * 1024 //100 Mb
	httpUserAgent                 = "reMP3 HTTP Fetcher"
)

var config = Config{}

func main() {
	parseFlags()

	listen, err := reuseport.Listen("tcp4", config.WebListen)
	if err != nil {
		log.Fatalf("Error in reuseport listener: %s", err)
	}

	router := getRouter()

	server := &fasthttp.Server{
		Handler:            router.Handler,
		DisableKeepalive:   true,
		MaxRequestBodySize: serverMaxBodySize,
		Concurrency:        serverMaxConcurrencyRequests,
		ReadTimeout:        serverRequestReadTimeout,
		WriteTimeout:       serverResponseWriteTimeout,
	}

	log.Printf("Server started on %s\n", config.WebListen)
	if err := server.Serve(listen); err != nil {
		log.Fatalf("Error in ListenAndServe: %s", err)
	}
}

func getRouter() *fasthttprouter.Router {
	router := fasthttprouter.New()
	router.GET("/*p", getResizeHandler)
	router.POST("/", postResizeHandler)
	return router
}

func parseFlags() {
	flag.StringVar(&config.WebListen, "WEB_LISTEN", "127.0.0.1:7090", "Listen interface and port")
	flag.StringVar(&config.FFmpegBinary, "CFG_FFMPEG_BINARY", "ffmpeg", "Path to ffmpeg binary")
	flag.DurationVar(&config.ServerConvertTimeout, "CFG_CONVERT_TIMEOUT", 30*time.Second, "Convert timeout.")
	flag.Parse()
}
