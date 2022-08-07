package torlook

import (
	"GetVideo/rutor/client"
	"GetVideo/torr"
	"encoding/json"
	"errors"
	"github.com/dustin/go-humanize"
	"log"
	"net/url"
	"time"
)

type Data struct {
	Date     int64  `json:"date,omitempty"`
	Leechers int    `json:"leechers,omitempty"`
	Magnet   string `json:"magnet,omitempty"`
	Seeders  int    `json:"seeders,omitempty"`
	Size     int64  `json:"size,omitempty"`
	Title    string `json:"title,omitempty"`
	TopicUrl string `json:"topic_url,omitempty"`
	Tracker  string `json:"tracker,omitempty"`
}

type TorlookJson struct {
	CachedData string `json:"cached_data,omitempty"`
	Data       []Data `json:"data,omitempty"`
	Error      bool   `json:"error,omitempty"`
	Message    string `json:"message,omitempty"`
}

func ParsePage(query string) ([]*torr.TorrentDetails, error) {
	var list []*torr.TorrentDetails

	body, err := get("https://api.torlook.info/api.php?key=ScYPftyYakbRRqw7&s=" + url.PathEscape(query))
	if err != nil {
		return nil, err
	}

	var tlJson *TorlookJson

	err = json.Unmarshal(body, &tlJson)
	if err != nil {
		return nil, err
	}

	if tlJson.Error {
		return nil, errors.New(tlJson.Message)
	}

	for _, tt := range tlJson.Data {
		if tt.Magnet == "" {
			continue
		}
		itm := new(torr.TorrentDetails)
		itm.Title = tt.Title
		itm.Magnet = tt.Magnet
		itm.Date = time.Unix(tt.Date, 0)
		itm.Peer = tt.Seeders
		itm.Seed = tt.Leechers
		itm.Size = humanize.Bytes(uint64(tt.Size))
		itm.Tracker = tt.Tracker
		itm.Link = "http://" + tt.Tracker + "/" + tt.TopicUrl
		list = append(list, itm)
	}

	return list, nil
}

func get(link string) ([]byte, error) {
	var body []byte
	var err error
	for i := 0; i < 1; i++ {
		body, err = client.GetBuf(link, "", "")
		if err == nil {
			break
		}
		log.Println("Error get page,tryes:", i+1, link)
		time.Sleep(time.Second * 2)
	}
	return body, err
}
