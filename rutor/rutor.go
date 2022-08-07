package rutor

import (
	"GetVideo/rutor/client"
	"GetVideo/torr"
	"bytes"
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
	"log"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var rutorHost = "http://rutor.info"

func ParsePage(query string) ([]*torr.TorrentDetails, error) {

	var list []*torr.TorrentDetails

	body, err := get(rutorHost + "/search/0/0/100/2/" + url.PathEscape(query))
	if err != nil {
		return nil, err
	}
	doc, err := goquery.NewDocumentFromReader(bytes.NewBufferString(body))
	if err != nil {
		return nil, err
	}

	doc.Find("div#index").Find("tr").Each(func(_ int, selection *goquery.Selection) {
		if selection.HasClass("backgr") {
			return
		}
		selTd := selection.Find("td")

		itm := new(torr.TorrentDetails)
		itm.Tracker = "rutor.info"
		itm.Date = parseDate(node2Text(selTd.Get(0)))
		itm.Title = node2Text(selTd.Get(1))
		itm.Magnet = selTd.Get(1).FirstChild.NextSibling.Attr[0].Val
		itm.Link = rutorHost + selTd.Get(1).LastChild.Attr[0].Val
		if len(selTd.Nodes) == 4 {
			itm.Size = node2Text(selTd.Get(2))
			peers := node2Text(selTd.Get(3))
			prarr := strings.Split(peers, "  ")
			if len(prarr) > 1 {
				itm.Seed, _ = strconv.Atoi(prarr[1])
				itm.Peer, _ = strconv.Atoi(prarr[0])
			}
		} else if len(selTd.Nodes) == 5 {
			itm.Size = node2Text(selTd.Get(3))
			peers := node2Text(selTd.Get(4))
			prarr := strings.Split(peers, "  ")
			if len(prarr) > 1 {
				itm.Seed, _ = strconv.Atoi(prarr[1])
				itm.Peer, _ = strconv.Atoi(prarr[0])
			}
		}

		list = append(list, itm)
	})

	return list, nil
}

func parseDate(date string) time.Time {
	var rutorMonth = map[string]int{
		"Янв": 1, "Фев": 2, "Мар": 3,
		"Апр": 4, "Май": 5, "Июн": 6,
		"Июл": 7, "Авг": 8, "Сен": 9,
		"Окт": 10, "Ноя": 11, "Дек": 12,
	}

	darr := strings.Split(date, " ")
	if len(darr) != 3 {
		return time.Date(0, 0, 0, 0, 0, 0, 0, time.Now().Location())
	}

	day, _ := strconv.Atoi(darr[0])
	month, _ := rutorMonth[darr[1]]
	year, _ := strconv.Atoi("20" + darr[2])

	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Now().Location())
}

func node2Text(node *html.Node) string {
	return strings.TrimSpace(strings.Replace((&goquery.Selection{Nodes: []*html.Node{node}}).Text(), "\u00A0", " ", -1))
}

func get(link string) (string, error) {
	var body string
	var err error
	for i := 0; i < 3; i++ {
		body, err = client.Get(link, "", "")
		if err == nil {
			break
		}
		log.Println("Error get page,tryes:", i+1, link)
		time.Sleep(time.Second * 2)
	}
	return body, err
}
