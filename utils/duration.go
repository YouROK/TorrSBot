package utils

import (
	"encoding/json"
	ffmpeg "github.com/u2takey/ffmpeg-go"
	"strconv"
)

func Trim(fileIn, fileOut string) error {
	return ffmpeg.Input(fileIn).
		Output(fileOut, ffmpeg.KwArgs{"ss": "00:00:30"},
			ffmpeg.KwArgs{"c": "copy"}).
		OverWriteOutput().ErrorToStdOut().Run()
}

func ProbeDuration(filename string) (float64, error) {
	a, err := ffmpeg.Probe(filename)
	if err != nil {
		return 0, err
	}
	return probeDuration(a)
}

type probeFormat struct {
	Duration string `json:"duration"`
}

type probeData struct {
	Format probeFormat `json:"format"`
}

func probeDuration(a string) (float64, error) {
	pd := probeData{}
	err := json.Unmarshal([]byte(a), &pd)
	if err != nil {
		return 0, err
	}
	f, err := strconv.ParseFloat(pd.Format.Duration, 64)
	if err != nil {
		return 0, err
	}
	return f, nil
}
