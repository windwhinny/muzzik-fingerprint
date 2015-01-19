package muzzikfp

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Muzzik-Dev-Group/muzzik-fingerprint/xiami"
	"io"
	"net/http"
	"net/url"
	"sort"
)

type Music struct {
	xiami.Music
	Score float32 `json:"score"`
}

type Matcher struct {
	FPs []string
	// 将 querySolr 分离出来, 以便做单元测试
	querySolr func(string) (Musics, error)
}

type Musics []*Music

type solrResponse struct {
	Response struct {
		Docs Musics `json:"docs"`
	} `json:"response"`
}

func (musics Musics) Len() int {
	return len(musics)
}

func (musics Musics) Swap(i, j int) {
	musics[i], musics[j] = musics[j], musics[i]
	return
}

func (musics Musics) Less(i, j int) bool {
	return musics[i].Score > musics[j].Score
}

// QuerySolr 通过转入的歌曲指纹字符串来查询对应的歌曲
func QuerySolr(fp string) (musics Musics, err error) {
	var query = url.Values{
		"q":    {fp},
		"rows": {"5"},
		"wt":   {"json"},
	}
	var link string
	if SolrHost == "" {
		link = `http://localhost:8080/solr/fp/select?fl=*%2Cscore`
	} else {
		link = fmt.Sprintf("http://%s/solr/fp/select?fl=*%2Cscore", SolrHost)
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
	musics = data.Response.Docs
	if len(musics) == 0 {
		err = errors.New("not found")
		return
	}
	return
}

// Match 将会根据 FPs 逐个请求，并返回最佳检索结果
func (matcher *Matcher) Match() (music *Music, err error) {
	var musics Musics
	if len(matcher.FPs) == 0 {
		err = errors.New("not fingerprint")
		return
	}

	for _, fp := range matcher.FPs {

		var _ms Musics
		_ms, err = matcher.querySolr(fp)
		if err != nil {
			return
		}

		// 合并查询结果
		for _, m1 := range _ms {
			found := false
			for _, m2 := range musics {
				if m1.Id == m2.Id {
					m2.Score += m1.Score
					found = true
					break
				}
			}

			if !found {
				musics = append(musics, m1)
			}
		}
	}

	sort.Sort(musics)

	music = musics[0]

	music.Score = music.Score / float32(len(matcher.FPs))

	return
}

// GetBestMatch 是 matcher.Match 的一个封装
func GetBestMatch(fps []string) (music *Music, err error) {
	for k, fp := range fps {
		fp, err = UncompressFP(fp)
		if err != nil {
			return
		}
		fps[k] = fp
	}
	matcher := &Matcher{FPs: fps}
	matcher.querySolr = QuerySolr
	music, err = matcher.Match()
	return
}
