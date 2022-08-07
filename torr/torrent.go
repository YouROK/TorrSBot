package torr

import (
	"GetVideo/settings"
	"GetVideo/torr/state"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type TorrentDetails struct {
	Title   string
	Size    string
	Date    time.Time
	Link    string
	Tracker string
	Peer    int
	Seed    int
	Magnet  string
}

func GetTorrentInfo(hash string) (*state.TorrentStatus, error) {
	host := settings.GetTSHost()
	link := host + "/stream?stat&link=" + url.QueryEscape(hash)
	resp, err := http.Get(link)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var ti *state.TorrentStatus

	err = json.Unmarshal(buf, &ti)
	return ti, err
}

func DownloadTorrentFile(dir, hash, id string) (string, error) {
	os.MkdirAll(filepath.Join(dir, hash), 0777)

	ti, err := GetTorrentInfo(hash)
	if err != nil {
		return "", err
	}

	idf, _ := strconv.Atoi(id)
	fst := ti.FindFile(idf)
	if fst == nil {
		return "", errors.New("file with " + id + " not found")
	}

	if fst.Length > 2*1024*1024*1024 {
		return "", errors.New("file size is bigger")
	}

	ext := filepath.Ext(fst.Path)
	basefn := strconv.Itoa(fst.Id)
	filename := filepath.Join(dir, ti.Hash, basefn+ext)

	if _, err = os.Stat(filename); err == nil {
		return filename, nil
	}

	host := settings.GetTSHost()
	link := host + "/stream?link=" + url.QueryEscape(hash) + "&index=" + id + "&play"
	resp, err := http.Get(link)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	ff, err := os.Create(filename)
	if err != nil {
		return "", err
	}
	defer ff.Close()
	_, err = io.Copy(ff, resp.Body)
	if err != nil {
		defer os.Remove(filename)
		return "", err
	}
	return filename, nil
}
