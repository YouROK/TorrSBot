package torr

import (
	"GetVideo/torr/state"
	"fmt"
	tele "gopkg.in/telebot.v3"
	"math"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

type DLQueue struct {
	id        int
	c         tele.Context
	hash      string
	fileID    string
	fileName  string
	updateMsg *tele.Message
}

var (
	queue   []*DLQueue
	mu, smu sync.Mutex
	isWork  bool
	idCount int
)

func Add(c tele.Context, hash, fileID string) {
	mu.Lock()
	defer mu.Unlock()
	idCount++
	if idCount > math.MaxInt {
		idCount = 0
	}
	dlQueue := &DLQueue{
		id:     idCount,
		c:      c,
		hash:   hash,
		fileID: fileID,
	}
	ti, _ := GetTorrentInfo(hash)
	if ti != nil {
		id, err := strconv.Atoi(dlQueue.fileID)
		if err == nil {
			file := ti.FindFile(id)
			if file != nil {
				dlQueue.fileName = file.Path
			}
		}
	}
	queue = append(queue, dlQueue)

	uMsg, _ := c.Bot().Send(c.Recipient(), "Подготовка к загрузке")

	dlQueue.updateMsg = uMsg
	go work()
	go sendStatus()
}

func Cancel(id int) {
	mu.Lock()
	defer mu.Unlock()
	for i, dlQueue := range queue {
		if dlQueue.id == id {
			dlQueue.c.Bot().Delete(dlQueue.updateMsg)
			queue = append(queue[:i], queue[i+1:]...)
			go sendStatus()
			return
		}
	}
}

func work() {
	smu.Lock()
	if isWork {
		smu.Unlock()
		return
	}
	isWork = true
	defer func() { isWork = false }()
	smu.Unlock()

	for true {
		mu.Lock()
		if len(queue) == 0 {
			mu.Unlock()
			break
		}
		dlQueue := queue[0]
		queue = queue[1:]
		mu.Unlock()
		sendStatus()

		ti, _ := GetTorrentInfo(dlQueue.hash)
		var file *state.TorrentFileStat
		if ti != nil {
			id, _ := strconv.Atoi(dlQueue.fileID)
			file = ti.FindFile(id)
		}

		dlQueue.c.Bot().Notify(dlQueue.c.Recipient(), tele.UploadingVideo)

		ext := strings.ToLower(filepath.Ext(file.Path))
		caption := filepath.Base(file.Path)
		torrFile, err := newTFile(dlQueue)
		if err != nil {
			dlQueue.c.Bot().Edit(dlQueue.updateMsg, err.Error())
			return
		}

		switch ext {
		case ".3g2", ".3gp", ".aaf", ".asf", ".avchd", ".avi", ".drc", ".flv", ".m2v", ".m3u8", ".m4p", ".m4v", ".mkv", ".mng", ".mov", ".mp2", ".mp4", ".mpe", ".mpeg", ".mpg", ".mpv", ".mxf", ".nsv", ".ogv", ".qt", ".rm", ".rmvb", ".roq", ".svi", ".vob", ".webm", ".wmv", ".yuv":
			{
				v := &tele.Video{}
				v.File.FileReader = torrFile
				v.FileName = file.Path
				v.Caption = caption
				err = dlQueue.c.Send(v)
			}
		case ".wav", ".bwf", ".raw", ".aiff", ".flac", ".m4a", ".pac", ".tta", ".wv", ".ast", ".aac", ".mp3", ".amr", ".s3m", ".act", ".au", ".dct", ".dss", ".gsm", ".mmf", ".mpc", ".ogg", ".oga", ".opus", ".ra", ".sln", ".vox":
			{
				a := &tele.Audio{}
				a.File.FileReader = torrFile
				a.FileName = file.Path
				a.Caption = caption
				err = dlQueue.c.Send(a)
			}
		default:
			d := &tele.Document{}
			d.File.FileReader = torrFile
			d.FileName = file.Path
			d.Caption = caption
			err = dlQueue.c.Send(d)
		}
		if err != nil {
			fmt.Println("Ошибка загрузки в телеграм:", err)
			errstr := fmt.Sprintf("Ошибка загрузки в телеграм: %v", file.Path)
			dlQueue.c.Bot().Edit(dlQueue.updateMsg, errstr)
			return
		}
		torrFile.Close()
		dlQueue.c.Bot().Delete(dlQueue.updateMsg)
	}
}

func sendStatus() {
	mu.Lock()
	defer mu.Unlock()
	for i, dlQueue := range queue {
		torrKbd := &tele.ReplyMarkup{}
		btnCancel := torrKbd.Data("Отмена", "downloadCancel", strconv.Itoa(dlQueue.id))
		rows := []tele.Row{torrKbd.Row(btnCancel)}
		torrKbd.Inline(rows...)

		msg := "Номер в очереди " + strconv.Itoa(i+1)
		if dlQueue.fileName != "" {
			msg += "\n<i>" + dlQueue.fileName + "</i>"
		}

		dlQueue.c.Bot().Edit(dlQueue.updateMsg, msg, torrKbd)
	}
}
