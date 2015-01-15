package xiami

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/windwhinny/muzzik-fingerprint/http-client"
	"html"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

var PathTmp = `http://www.xiami.com/song/playlist/id/%d/cat/json`

type Id int

type Music struct {
	Title       string `json:"title,omitempty"`
	Album       string `json:"album,omitempty"`
	Artist      string `json:"artist,omitempty"`
	Cover       string `json:"cover,omitempty"`
	XiamiId     Id     `json:"xiamiId,omitempty"`
	Id          string `json:"id,omitempty"`
	FingerPrint string `json:"fp,omitempty"`
	Url         string `json:"-"`
}

type XiamiResponse struct {
	Data struct {
		TrackList []XiamiTrack `json:"trackList,omitempty"`
		Uid       string       `json:"uid,omitempty"`
	} `json:"data,omitempty"`
}

type XiamiTrack struct {
	Title      string `json:"title,omitempty"`
	Album      string `json:"album_name,omitempty"`
	Artist     string `json:"artist,omitempty"`
	Cover      string `json:"album_pic,omitempty"`
	Url        string `json:"location,omitempty"`
	Id         string `json:"song_id,omitempty"`
	urlDecoded bool
}

func GetMusic(id Id) (music *Music, err error) {
	var res *http.Response

	link := fmt.Sprintf(PathTmp, id)
	res, err = httpClient.Get(link)
	if err != nil {
		return
	}
	defer res.Body.Close()
	var buf []byte
	xiamiRes := &XiamiResponse{}
	buf, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}
	err = json.Unmarshal(buf, xiamiRes)
	if err != nil {
		return
	}

	if len(xiamiRes.Data.TrackList) <= 0 {
		err = errors.New("empty response")
		return
	}

	track := xiamiRes.Data.TrackList[0]

	err = track.DecodeUrl()
	if err != nil {
		return
	}

	music, err = track.ConvertMusic()
	return
}

func (track *XiamiTrack) DecodeUrl() (err error) {
	if track.urlDecoded {
		return
	}

	if track.Url == `` {
		return
	}

	var data []string
	rows := int([]byte(track.Url)[0] - 48)
	str := track.Url[1:]
	colums := int(math.Ceil(float64(len(str)) / float64(rows)))
	check := `ttp%3A%2F`
	_url := ``

	for i := 0; i < rows; i++ {
		l := colums
		if len(str) >= l {
			if str[l-1] == check[i] {
				l -= 1
			}
			data = append(data, str[0:l])
			str = str[l:]
		} else {
			data = append(data, str)
			break
		}
	}

	for i := 0; i < colums; i++ {
		_url += func(i int) string {
			var result []byte
			for _, str := range data {
				if len(str) > i {
					result = append(result, str[i])
				} else {
					break
				}
			}

			return string(result)
		}(i)
	}

	_url, err = url.QueryUnescape(_url)

	if err != nil {
		return
	}

	_url = strings.Replace(_url, "^", "0", -1)
	track.Url = _url
	track.urlDecoded = true
	return
}

func (track *XiamiTrack) ConvertMusic() (music *Music, err error) {
	music = &Music{}
	music.Title = html.UnescapeString(track.Title)
	music.Album = html.UnescapeString(track.Album)
	music.Artist = html.UnescapeString(track.Artist)
	music.Cover = track.Cover
	music.Url = track.Url

	var xiamiId int64
	xiamiId, err = strconv.ParseInt(track.Id, 10, 0)

	music.XiamiId = Id(xiamiId)
	return
}
