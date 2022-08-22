package main

import (
	"GetVideo/rutor"
	"GetVideo/settings"
	"GetVideo/torlook"
	"GetVideo/torr"
	"GetVideo/utils"
	"errors"
	"fmt"
	"github.com/dustin/go-humanize"
	tele "gopkg.in/telebot.v3"
	"log"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func main() {

	pref := tele.Settings{
		URL:       settings.GetTGHost(),
		Token:     settings.GetTGBotApi(),
		Poller:    &tele.LongPoller{Timeout: 60 * time.Second},
		ParseMode: tele.ModeHTML,
	}

	b, err := tele.NewBot(pref)
	if err != nil {
		log.Fatal(err)
		return
	}

	b.Handle("help", help)
	b.Handle("Help", help)
	b.Handle("/help", help)
	b.Handle("/Help", help)
	b.Handle("/start", help)

	b.Handle(tele.OnText, func(c tele.Context) error {
		txt := c.Text()
		if strings.HasPrefix(strings.ToLower(txt), "magnet:") || utils.IsHash(txt) {
			return infoTorrent(c, c.Text())
		} else if strings.HasPrefix(strings.ToLower(txt), "get:") {
			return getTorrent(c)
		} else {
			return findTorrents(c, c.Text())
		}
	})

	b.Handle(tele.OnCallback, func(c tele.Context) error {
		args := c.Args()
		if len(args) > 0 {
			if args[0] == "\ffile" {
				return getTorrent(c)
			}
			if args[0] == "\ftorr" {
				return infoTorrent(c, args[1])
			}
			if args[0] == "\fdownloadCancel" {
				if num, err := strconv.Atoi(args[1]); err == nil {
					torr.Cancel(num)
					return nil
				}
			}
		}
		return errors.New("Ошибка кнопка не распознана")
	})

	b.Start()
}

func help(c tele.Context) error {
	return c.Send("Для поиска по рутор введите название фильма или сериала в конце можно дописать год, качество, релизера\n" +
		"Пример: <i>бэтмен 2022 кпк</i>\n" +
		"Чем больше слов, тем лучше результат\n\n" +
		"Для скачивания можно вставить магнет или хэш торрента\n" +
		"Лимит телеграма на загружаемый файл 2гб, выбирайте торренты, где файл будет меньше 2гб")
}

func findTorrents(c tele.Context, query string) error {
	list, err := rutor.ParsePage(query)
	if err != nil {
		list, err = torlook.ParsePage(query)
		if err != nil {
			c.Send(err.Error())
			return err
		}
	}

	if len(list) == 0 {
		c.Send("Торрент не найден")
		return nil
	}

	for _, d := range list {
		txt := fmt.Sprintf("<b>%v</b>\n%v <b>%v</b> ↑%v ↓%v %v", d.Title, d.Date.Format("01.02.2006"), d.Size, d.Peer, d.Seed, d.Tracker)
		mag, err := url.Parse(d.Magnet)
		if err != nil {
			fmt.Println("Ошибка в магнет ссылке:", d.Magnet, err)
			continue
		}
		arr := strings.Split(mag.Query().Get("xt"), ":")
		if len(arr) != 3 {
			fmt.Println("Ошибка в магнет ссылке:", d.Magnet)
			continue
		}
		torrKbd := &tele.ReplyMarkup{}
		btnDwn := torrKbd.Data("Загрузить", "torr", arr[2])
		btnLnk := torrKbd.URL("Ссылка", d.Link)
		rows := []tele.Row{torrKbd.Row(btnDwn), torrKbd.Row(btnLnk)}
		torrKbd.Inline(rows...)
		c.Send(txt, torrKbd)
	}

	return nil
}

func infoTorrent(c tele.Context, magnet string) error {
	ti, err := torr.GetTorrentInfo(magnet)
	if err != nil {
		return c.Send(err.Error())
	}
	txt := "<b>" + ti.Title + "</b>\n" +
		"<code>" + ti.Hash + "</code>"

	filesKbd := &tele.ReplyMarkup{}
	var files []tele.Row

	i := len(txt)
	for _, f := range ti.FileStats {
		btn := filesKbd.Data(filepath.Base(f.Path)+" "+humanize.Bytes(uint64(f.Length)), "file", ti.Hash, strconv.Itoa(f.Id))
		files = append(files, filesKbd.Row(btn))
		if i+len(txt) > 4096 {
			filesKbd := &tele.ReplyMarkup{}
			filesKbd.Inline(files...)
			c.Send(txt, filesKbd)
			files = files[:0]
			i = len(txt)
		}
		i += len(filepath.Base(f.Path) + " " + humanize.Bytes(uint64(f.Length)))
	}
	filesKbd.Inline(files...)

	return c.Send(txt, filesKbd)
}

func getTorrent(c tele.Context) error {
	args := c.Args()
	if len(args) != 3 {
		return errors.New("Ошибка не верные данные")
	}

	torr.Add(c, args[1], args[2])

	return nil
}
