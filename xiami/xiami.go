package xiami

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Muzzik-Dev-Group/muzzik-fingerprint/http-client"
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
	Id          string `json:"id,omitempty"`
	Title       string `json:"title,omitempty"`
	Album       string `json:"album,omitempty"`
	Artist      string `json:"artist,omitempty"`
	Cover       string `json:"cover,omitempty"`
	XiamiId     Id     `json:"xiamiId,omitempty"`
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

// GetMusic 通过传入的虾米歌曲 Id 返回对应的 Music 实例
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

// DecodeUrl 用来解密虾米返回的 mp3 地址。
// 其地址加密方式类似凯撒矩阵，如下:
//
// <code>6hAFlm%%7422758.Fk459d58642Ent%mei22562F%74maedE98E5%2%-ut25..FF22615_2puy5e%b24515%lpF.xc71%%97E413t%3d5bcbE5E5l%%fio52256782_%h3f5Eb9d-3%E32iam25FE%139l3_D%4f%6d195-</code>
//
// 转换后
//
// <code>
// 6
// hAFlm%%7422758.Fk459d58642En
// t%mei22562F%74maedE98E5%2%-u
// t25..FF22615_2puy5e%b24515%l
// pF.xc71%%97E413t%3d5bcbE5E5l
// %%fio52256782_%h3f5Eb9d-3%E
// 32iam25FE%139l3_D%4f%6d195-
// <code>
// 其第一位为行数
//
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

// ConvertMusic 将虾米的歌曲结构转换为我们自己的结构
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
