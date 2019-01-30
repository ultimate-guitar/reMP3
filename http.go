package main

import (
	"fmt"
	"github.com/labstack/gommon/log"
	"github.com/valyala/fasthttp"
	"io"
	"io/ioutil"
	"net/http"
)

type requestParams struct {
	fileUrl         fasthttp.URI
	mp3Original     []byte
	mp3Resized      []byte
	reBitrate       int
	reDuration      int
	reOutSampleRate int
}

var httpTransport = &http.Transport{
	MaxIdleConns:        httpClientMaxIdleConns,
	IdleConnTimeout:     httpClientIdleConnTimeout,
	MaxIdleConnsPerHost: httpClientMaxIdleConnsPerHost,
}
var httpClient = &http.Client{Transport: httpTransport, Timeout: httpClientFileDownloadTimeout}

func getResizeHandler(ctx *fasthttp.RequestCtx) {
	var code int
	var err error
	params := requestParams{}
	if err := getRequestParser(ctx, &params); err != nil {
		log.Printf("Can not parse requested url: '%s', err: %s", ctx.URI(), err)
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}
	params.mp3Original, code, err = getSourceFile(params.fileUrl)
	if err != nil {
		log.Printf("Can not get source file: '%s', err: %s", params.fileUrl.String(), err)
		ctx.SetStatusCode(code)
		return
	}

	if err := resizeMP3(&params); err != nil {
		log.Printf("Can not resize file: '%s', err: %s", params.fileUrl.String(), err)
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		return
	}
	ctx.SetBody(params.mp3Resized)
	ctx.SetContentType("audio/mpeg")
	ctx.SetStatusCode(fasthttp.StatusOK)
	return
}

func postResizeHandler(ctx *fasthttp.RequestCtx) {
	params := requestParams{}
	if err := postRequestParser(ctx, &params); err != nil {
		log.Printf("Can not parse requested url: '%s', err: %s", ctx.URI(), err)
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}

	params.mp3Original = ctx.PostBody()
	if err := resizeMP3(&params); err != nil {
		log.Printf("Can not resize file, err: %s", err)
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		return
	}
	ctx.SetBody(params.mp3Resized)
	ctx.SetContentType("audio/mpeg")
	ctx.SetStatusCode(fasthttp.StatusOK)
	return
}

func getRequestParser(ctx *fasthttp.RequestCtx, params *requestParams) (err error) {
	ctx.URI().CopyTo(&params.fileUrl)
	// Cleanup file URL from resize args
	for _, arg := range []string{resizeQArgBitrateName, resizeQArgOutSampleRateName, resizeQArgDurationName} {
		params.fileUrl.QueryArgs().Del(arg)
	}

	sourceHeader := string(ctx.Request.Header.Peek(resizeHeaderNameSource))
	if (sourceHeader == "") && ctx.IsGet() {
		return fmt.Errorf("empty '%s' header on GET request", resizeHeaderNameSource)
	}
	params.fileUrl.SetHost(sourceHeader)

	switch schemaHeader := string(ctx.Request.Header.Peek(resizeHeaderNameSchema)); schemaHeader {
	case "":
		params.fileUrl.SetScheme(resizeHeaderDefaultSchema)
	case "https":
		params.fileUrl.SetScheme("https")
	case "http":
		params.fileUrl.SetScheme("http")
	default:
		return fmt.Errorf("wrong '%s' header value: '%s'", resizeHeaderNameSchema, schemaHeader)
	}

	if ctx.URI().QueryArgs().Has(resizeQArgBitrateName) {
		params.reBitrate, err = ctx.URI().QueryArgs().GetUint(resizeQArgBitrateName)
		if err != nil {
			return fmt.Errorf("invalid bitrate value: %s", ctx.URI().QueryArgs().Peek(resizeQArgBitrateName))
		}
	} else {
		return fmt.Errorf("invalid request, no '%s' query param provided", resizeQArgBitrateName)
	}

	if ctx.URI().QueryArgs().Has(resizeQArgOutSampleRateName) {
		params.reOutSampleRate, err = ctx.URI().QueryArgs().GetUint(resizeQArgOutSampleRateName)
		if err != nil {
			return fmt.Errorf("invalid samplerate value: %s", ctx.URI().QueryArgs().Peek(resizeQArgOutSampleRateName))
		}
	} else {
		return fmt.Errorf("invalid request, no '%s' query param provided", resizeQArgOutSampleRateName)
	}

	params.reDuration, err = ctx.URI().QueryArgs().GetUint(resizeQArgDurationName)
	if err != nil {
		params.reDuration = 0
	}

	return nil
}

func postRequestParser(ctx *fasthttp.RequestCtx, params *requestParams) (err error) {
	if ctx.URI().QueryArgs().Has(resizeQArgBitrateName) {
		params.reBitrate, err = ctx.URI().QueryArgs().GetUint(resizeQArgBitrateName)
		if err != nil {
			return fmt.Errorf("invalid bitrate value: %s", ctx.URI().QueryArgs().Peek(resizeQArgBitrateName))
		}
	} else {
		return fmt.Errorf("invalid request, no '%s' query param provided", resizeQArgBitrateName)
	}

	if ctx.URI().QueryArgs().Has(resizeQArgOutSampleRateName) {
		params.reOutSampleRate, err = ctx.URI().QueryArgs().GetUint(resizeQArgOutSampleRateName)
		if err != nil {
			return fmt.Errorf("invalid samplerate value: %s", ctx.URI().QueryArgs().Peek(resizeQArgOutSampleRateName))
		}
	} else {
		return fmt.Errorf("invalid request, no '%s' query param provided", resizeQArgOutSampleRateName)
	}

	params.reDuration, err = ctx.URI().QueryArgs().GetUint(resizeQArgDurationName)
	if err != nil {
		params.reDuration = 0
	}

	return nil
}

func getSourceFile(url fasthttp.URI) (data []byte, code int, err error) {
	client := new(http.Client)
	req, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		return data, fasthttp.StatusInternalServerError, err
	}

	req.Header.Set("User-Agent", httpUserAgent)
	res, err := client.Do(req)
	if res != nil {
		defer res.Body.Close()
		defer io.Copy(ioutil.Discard, res.Body)
	}
	if err != nil {
		return data, fasthttp.StatusInternalServerError, err
	}

	if res.StatusCode != fasthttp.StatusOK {
		return data, res.StatusCode, fmt.Errorf("status code %d != %d", res.StatusCode, fasthttp.StatusOK)
	}

	data, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return data, fasthttp.StatusInternalServerError, err
	}
	return data, res.StatusCode, nil
}
