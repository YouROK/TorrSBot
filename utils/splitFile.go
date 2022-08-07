package utils

import (
	ffmpeg "github.com/u2takey/ffmpeg-go"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func SplitFile(inputFile string, sizeLimit int64) ([]string, error) {
	currDuration := 0.0
	dir := filepath.Dir(inputFile)
	ext := filepath.Ext(inputFile)
	file := strings.TrimSuffix(filepath.Base(inputFile), ext)
	index := 1
	filePart := filepath.Join(dir, file) + "." + strconv.Itoa(index) + ext

	totalDuration, err := ProbeDuration(inputFile)
	if err != nil {
		return nil, err
	}
	time.Sleep(time.Second * 2)
	var files []string
	for int(currDuration) < int(totalDuration) {
		ffmpeg.Input(inputFile).
			Output(filePart, ffmpeg.KwArgs{"ss": currDuration}, ffmpeg.KwArgs{"fs": sizeLimit}, ffmpeg.KwArgs{"c": "copy"}).
			OverWriteOutput().Run()
		newDuration, err := ProbeDuration(filePart)
		if err != nil {
			return nil, err
		}
		files = append(files, filePart)
		currDuration += newDuration
		index += 1
		filePart = filepath.Join(dir, file) + "." + strconv.Itoa(index) + ext
	}
	return files, nil
}
