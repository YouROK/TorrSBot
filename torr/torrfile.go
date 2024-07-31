package torr

import (
	"GetVideo/settings"
	"errors"
	"fmt"
	"github.com/dustin/go-humanize"
	tele "gopkg.in/telebot.v3"
	"net"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"time"
)

type TorrFile struct {
	hash    string
	name    string
	dlQueue *DLQueue
	offset  int64
	size    int64
	id      int

	lastUpdate time.Time
	resp       *http.Response
}

func newTFile(dlQueue *DLQueue) (*TorrFile, error) {
	hash := dlQueue.hash
	fileID := dlQueue.fileID
	dlQueue.c.Bot().Edit(dlQueue.updateMsg, "Подключение к торренту")
	ti, err := GetTorrentInfo(hash)
	if err != nil {
		return nil, err
	}

	id, _ := strconv.Atoi(fileID)
	tfile := ti.FindFile(id)
	if tfile == nil {
		return nil, errors.New("Файл с id:" + fileID + "в торренте не найден")
	}

	if tfile.Length > 2*1024*1024*1024 {
		return nil, errors.New("Размер файла должен быть меньше 2GB")
	}

	tf := new(TorrFile)
	tf.hash = hash
	tf.name = filepath.Base(tfile.Path)
	tf.dlQueue = dlQueue
	tf.id = id
	tf.size = tfile.Length

	host := settings.GetTSHost()
	link := host + "/stream?link=" + url.QueryEscape(hash) + "&index=" + fileID + "&play"
	c := &http.Client{
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).Dial,
			TLSHandshakeTimeout:   10 * time.Second,
			ResponseHeaderTimeout: 10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}

	resp, err := c.Get(link)
	if err != nil {
		return nil, err
	}
	tf.resp = resp

	return tf, nil
}

func (t *TorrFile) Read(p []byte) (n int, err error) {
	n, err = t.resp.Body.Read(p)
	if err == nil {
		t.offset += int64(n)
		since := time.Since(t.lastUpdate).Seconds()
		if since > 1.0 || t.offset >= t.size {
			t.UpdateStatus()
			t.dlQueue.c.Bot().Notify(t.dlQueue.c.Recipient(), tele.UploadingVideo)
		}
	}
	return
}

func (t *TorrFile) UpdateStatus() {
	ti, err := GetTorrentInfo(t.hash)
	if err != nil {
		t.dlQueue.c.Bot().Edit(t.dlQueue.updateMsg, "Ошибка при получении данных о торренте")
	} else {
		wait := time.Duration(float64(t.size-t.offset)/ti.DownloadSpeed) * time.Second
		speed := humanize.Bytes(uint64(ti.DownloadSpeed)) + "/sec"
		//speed := humanize.Bytes(uint64(t.speed)) + "/sec"
		peers := fmt.Sprintf("%v · %v/%v", ti.ConnectedSeeders, ti.ActivePeers, ti.TotalPeers)
		prc := fmt.Sprintf("%.2f%% %v / %v", float64(t.offset)*100.0/float64(t.size), humanize.Bytes(uint64(t.offset)), humanize.Bytes(uint64(t.size)))

		name := t.name
		if name == ti.Title {
			name = ""
		}

		msg := "Загрузка торрента:\n" +
			"<b>" + ti.Title + "</b>\n"
		if name != "" {
			msg += "<i>" + name + "</i>\n"
		}
		msg += "<b>Хэш: </b><code>" + t.hash + "</code>\n" +
			"<b>Скорость: </b>" + speed + "\n" +
			"<b>Осталось: </b>" + wait.String() + "\n" +
			"<b>Пиры: </b>" + peers + "\n" +
			"<b>Загружено: </b>" + prc

		if t.offset >= t.size {
			msg += "\n<b>Завершение загрузки</b>"
		}

		t.dlQueue.c.Bot().Edit(t.dlQueue.updateMsg, msg)
	}
	t.lastUpdate = time.Now()
}

func (t *TorrFile) Close() {
	if t.resp != nil && t.resp.Body != nil {
		t.resp.Body.Close()
		t.resp = nil
	}
}
