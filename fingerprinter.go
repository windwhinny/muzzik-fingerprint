package muzzikfp

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/windwhinny/muzzik-fingerprint/xiami"
	"io"
	"net/http"
	"net/url"
)

type Scanner struct {
	Filename string
	Music    *xiami.Music
}

type solrResponse struct {
	Response struct {
		Docs []*xiami.Music `json:"docs"`
	} `json:"response"`
}

func querySolr(fp string) (music *xiami.Music, err error) {
	var query = url.Values{
		"q":    {fp},
		"rows": {"1"},
		"wt":   {"json"},
	}
	var link string
	if SolrHost == "" {
		link = `http://localhost:8080/solr/fp/select`
	} else {
		link = fmt.Sprintf("http://%s/solr/fp/select", SolrHost)
	}
	res, err := http.PostForm(link, query)
	if err != nil {
		return
	}
	defer res.Body.Close()
	var buf bytes.Buffer
	io.Copy(&buf, res.Body)
	data := &solrResponse{}
	err = json.Unmarshal(buf.Bytes(), data)
	if err != nil {
		return
	}
	docs := data.Response.Docs
	if len(docs) == 0 {
		err = errors.New("not found")
		return
	}
	music = docs[0]
	return
}

func (scanner *Scanner) Match() (err error) {
	var fp string
	var music *xiami.Music
	if scanner.Filename == "" {
		err = errors.New("filename not set")
		return
	}
	fp, err = getRangeFingerPrint(scanner.Filename, 0, 10)
	if err != nil {
		return
	}

	music, err = querySolr(fp)
	if err != nil {
		return
	}

	scanner.Music = music
	return
}
