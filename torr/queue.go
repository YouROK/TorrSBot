package torr

import (
	"GetVideo/settings"
	"GetVideo/torr/state"
	"fmt"
	"github.com/dustin/go-humanize"
	tele "gopkg.in/telebot.v3"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
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

		isDownload := true
		caption := ""
		dir := settings.GetDownloadDir()

		go func() {
			for isDownload {
				ti, _ := GetTorrentInfo(dlQueue.hash)
				if ti != nil {
					id, _ := strconv.Atoi(dlQueue.fileID)
					file := ti.FindFile(id)
					if caption == "" {
						caption = file.Path
					}
					if file != nil {
						speed := humanize.Bytes(uint64(ti.DownloadSpeed)) + "/sec"
						peers := fmt.Sprintf("%v · %v/%v", ti.ConnectedSeeders, ti.ActivePeers, ti.TotalPeers)
						prc := fmt.Sprintf("%.2f %%", float64(ti.BytesReadData)*100.0/float64(file.Length))
						dlQueue.c.Bot().Edit(dlQueue.updateMsg, "Загрузка торрента:\n"+
							"<b>"+ti.Title+"</b>\n"+
							"<i>"+file.Path+"</i>\n"+
							"<b>"+ti.Hash+"</b>\n"+
							"<b>Размер: </b> "+humanize.Bytes(uint64(file.Length))+"\n"+
							"<b>Скорость: </b>"+speed+"\n"+
							"<b>Пиры: </b>"+peers+"\n"+
							"<b>Загружено: </b>"+prc,
						)
					}
				}
				time.Sleep(time.Second)
			}
		}()
		filePath, err := DownloadTorrentFile(dir, dlQueue.hash, dlQueue.fileID)
		isDownload = false
		defer func() {
			os.RemoveAll(filepath.Join(dir, dlQueue.hash))
		}()
		if err != nil {
			dlQueue.c.Bot().Edit(dlQueue.updateMsg, err.Error())
			return
		}

		ti, _ := GetTorrentInfo(dlQueue.hash)
		var file *state.TorrentFileStat
		if ti != nil {
			id, _ := strconv.Atoi(dlQueue.fileID)
			file = ti.FindFile(id)
		}

		ext := strings.ToLower(filepath.Ext(file.Path))

		dlQueue.c.Bot().Edit(dlQueue.updateMsg, "Загрузка в телеграм...")
		switch ext {
		case ".3g2", ".3gp", ".aaf", ".asf", ".avchd", ".avi", ".drc", ".flv", ".m2v", ".m3u8", ".m4p", ".m4v", ".mkv", ".mng", ".mov", ".mp2", ".mp4", ".mpe", ".mpeg", ".mpg", ".mpv", ".mxf", ".nsv", ".ogv", ".qt", ".rm", ".rmvb", ".roq", ".svi", ".vob", ".webm", ".wmv", ".yuv":
			{
				v := &tele.Video{}
				v.File.FileReader = newSFile(filePath, dlQueue)
				v.FileName = file.Path
				v.Caption = caption
				err = dlQueue.c.Send(v)
			}
		case ".wav", ".bwf", ".raw", ".aiff", ".flac", ".m4a", ".pac", ".tta", ".wv", ".ast", ".aac", ".mp3", ".amr", ".s3m", ".act", ".au", ".dct", ".dss", ".gsm", ".mmf", ".mpc", ".ogg", ".oga", ".opus", ".ra", ".sln", ".vox":
			{
				a := &tele.Audio{}
				a.File.FileReader = newSFile(filePath, dlQueue)
				a.FileName = file.Path
				a.Caption = caption
				err = dlQueue.c.Send(a)
			}
		default:
			d := &tele.Document{}
			d.File.FileReader = newSFile(filePath, dlQueue)
			d.FileName = file.Path
			d.Caption = caption
			err = dlQueue.c.Send(d)
		}
		if err != nil {
			errstr := fmt.Sprintf("Ошибка загрузки в телеграм: %v %v", filePath, err)
			dlQueue.c.Bot().Edit(dlQueue.updateMsg, errstr)
			return
		}
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

type SendFile struct {
	*os.File
	dl     *DLQueue
	offset int
}

func newSFile(path string, dlQueue *DLQueue) *SendFile {
	ff, err := os.Open(path)
	if err == nil {
		return &SendFile{
			File: ff,
			dl:   dlQueue,
		}
	}
	return nil
}

func (s *SendFile) Read(p []byte) (n int, err error) {
	n, err = s.File.Read(p)
	if err == nil {
		s.offset += n
		s.dl.c.Bot().Edit(s.dl.updateMsg, "Загрузка в телеграм:\n"+
			"<b>Загружено: </b>"+humanize.Bytes(uint64(s.offset)),
		)
	}
	return
}
