package main

import (
	"fmt"
	"github.com/labstack/gommon/log"
	"os/exec"
	"strconv"
	"syscall"
	"time"
)
import (
	"bytes"
)

type popenInput struct {
	args    []string
	stdin   *bytes.Buffer
	timeout time.Duration
}

type popenOutput struct {
	exitCode int
	stdout   bytes.Buffer
	stderr   bytes.Buffer
}

var ffmpegError = fmt.Errorf("ffmpeg command failed")

func resizeMP3(params *requestParams) error {
	input := popenInput{stdin: bytes.NewBuffer(params.mp3Original), timeout: config.ServerConvertTimeout}

	//Set input
	input.args = []string{"-f", "mp3", "-i", "pipe:"}

	//Bitrate filter
	input.args = append(input.args, "-b:a", strconv.Itoa(params.reBitrate))

	//SampleRate filter
	input.args = append(input.args, "-ar", strconv.Itoa(params.reOutSampleRate))

	//Remove all except audio
	input.args = append(input.args, "-map", "a")

	//Crop
	if params.reDuration > 0 {
		input.args = append(input.args, "-t", strconv.Itoa(params.reDuration))
	}

	//Set output
	input.args = append(input.args, "-f", "mp3", "pipe:")

	out, err := ffMpegPopen(&input)
	if err != nil {
		return err
	}
	if out.exitCode != 0 {
		log.Infof("ffmpeg with args: '%s' failed with exit code: %d", input.args, out.exitCode)
		return ffmpegError
	}

	params.mp3Resized = out.stdout.Bytes()
	return nil
}

func ffMpegPopen(input *popenInput) (output popenOutput, err error) {
	var timer *time.Timer

	cmd := exec.Command(config.FFmpegBinary, input.args...)

	cmd.Stdin = input.stdin
	cmd.Stdout = &output.stdout
	cmd.Stderr = &output.stderr

	err = cmd.Start()
	if err != nil {
		return
	}

	timer = time.AfterFunc(input.timeout, func() {
		if input.timeout > time.Duration(0) {
			cmd.Process.Kill()
		}
	})

	err = cmd.Wait()

	timer.Stop()

	if err != nil {
		exitError := err.(*exec.ExitError)
		ws := exitError.Sys().(syscall.WaitStatus)
		output.exitCode = ws.ExitStatus()
	} else {
		ws := cmd.ProcessState.Sys().(syscall.WaitStatus)
		output.exitCode = ws.ExitStatus()
	}
	return output, nil
}
