package main

import (
	"encoding/json"
	"errors"
	"github.com/PuerkitoBio/goquery"
	"github.com/iand/youtube"
	"io"
	"net/url"
	"strings"
)

type SearchResult struct {
	Title           string
	DurationSeconds int
	Url             string
}

func Search(q string) ([]SearchResult, error) {
	c := youtube.New()
	f, e := c.VideoSearch(q)
	if e != nil {
		return nil, e
	}
	res := make([]SearchResult, len(f.Entries))
	for i, entry := range f.Entries {
		title := entry.Title.Value
		dur := entry.Media.Duration.Seconds
		var link string
		for _, l := range entry.Links {
			if l.Rel == "alternate" {
				link = l.Href
				break
			}
		}
		res[i] = SearchResult{title, dur, link}
	}
	return res, nil
}

type DownloadResult struct {
	Url       string
	Mime      string
	Extension string
}

func ParseDownload(page io.Reader) (*DownloadResult, error) {
	d, e := goquery.NewDocumentFromReader(page)
	if e != nil {
		return nil, e
	}
	src := d.Find("#player-api").Next().Next().Text()
	if len(src) < 100 {
		return nil, errors.New("Too short ytplayer.config")
	}
	src = src[48 : len(src)-1]
	var cfg ytconfig
	e = json.Unmarshal([]byte(src), &cfg)
	if e != nil {
		return nil, e
	}
	score := 0
	var dr DownloadResult
	ss := strings.Split(cfg.Args.StreamMap, `,`)
	for _, s := range ss {
		sm, e := url.ParseQuery(s)
		if e != nil {
			return nil, e
		}
		myscore := qmap[sm.Get(`quality`)] + tqmap[sm.Get(`type`)]
		if myscore >= score {
			score = myscore
			dr.Url = sm.Get(`url`) + `&signature=` + url.QueryEscape(sm.Get(`sig`))
			dr.Mime = strings.SplitN(sm.Get(`type`), `;`, 2)[0]
			dr.Extension = tmap[dr.Mime]
		}
	}
	if score == 0 {
		return nil, errors.New("No suitable video")
	}
	return &dr, nil
}

var tmap = map[string]string{
	`video/webm`:  `webm`,
	`video/mp4`:   `mp4`,
	`video/3gpp`:  `3gpp`,
	`video/x-flv`: `flv`,
}
var tqmap = map[string]int{
	`video/webm`:  40,
	`video/mp4`:   30,
	`video/3gpp`:  20,
	`video/x-flv`: 10,
}

var qmap = map[string]int{
	`small`:   100,
	`medium`:  200,
	`large`:   300,
	`hd720`:   400,
	`hd1080`:  500,
	`highres`: 600,
}

type ytconfig struct {
	Args args `json:"args"`
}

type args struct {
	StreamMap string `json:"url_encoded_fmt_stream_map"`
}
