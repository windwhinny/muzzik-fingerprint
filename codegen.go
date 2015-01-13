package muzzikfp

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/windwhinny/muzzik-fingerprint/xiami"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type FPWorkFlow struct {
	Music    *xiami.Music
	Filename string
}

type JSON map[string]interface{}

func updateSolr(data JSON) (err error) {
	var buf bytes.Buffer
	err = json.NewEncoder(&buf).Encode(data)
	if err != nil {
		return
	}

	var res *http.Response
	res, err = http.Post("http://localhost:8080/solr/fp/update", "Content-type:application/json", &buf)
	if err != nil {
		return
	}
	res.Body.Close()
	return
}

func (wf *FPWorkFlow) Save() (err error) {
	if wf.Music == nil {
		err = errors.New("wf have no music")
		return
	}

	update := JSON{
		"add": JSON{
			"doc": wf.Music,
		},
	}
	err = updateSolr(update)
	return
}

func (wf *FPWorkFlow) Remove() (err error) {
	if wf.Music == nil {
		err = errors.New("wf have no music")
		return
	}
	update := JSON{
		"delete": JSON{
			"id": wf.Music.Id,
		},
	}
	err = updateSolr(update)
	return
}

func (wf *FPWorkFlow) SetMusic(id xiami.Id) (err error) {
	var name, fp string
	var music *xiami.Music
	music, name, err = download(id)
	if err != nil {
		return err
	}
	wf.Filename = name
	fp, err = getFingerPrint(name)
	if err != nil {
		return err
	}

	music.FingerPrint = fp
	music.XiamiId = id
	music.Id = strconv.Itoa(int(id))
	wf.Music = music
	return
}

func download(id xiami.Id) (music *xiami.Music, name string, err error) {
	var res *http.Response
	music, err = xiami.GetMusic(id)
	if err != nil {
		return
	}
	res, err = http.Get(music.Url)

	if err != nil {
		return
	}

	var file *os.File
	file, err = ioutil.TempFile("", "muzzikfp")
	if err != nil {
		return
	}

	defer file.Close()

	_, err = io.Copy(file, res.Body)

	if err != nil {
		return
	}

	name = file.Name()

	return
}

func getRangeFingerPrint(file string, star int, end int) (fp string, err error) {
	var cmd *exec.Cmd
	if star < 0 || end < 0 {
		cmd = exec.Command("echoprint-codegen", file)
	} else {
		cmd = exec.Command("echoprint-codegen", file, strconv.Itoa(star), strconv.Itoa(end))
	}
	var buf bytes.Buffer
	cmd.Stdout = &buf

	err = cmd.Run()
	if err != nil {
		return
	}
	var m []*struct {
		Code string `json:"code"`
	}

	err = json.Unmarshal(buf.Bytes(), &m)
	if err != nil {
		return
	}

	if len(m) == 0 {
		err = errors.New(fmt.Sprintf("cant get fingerprint of file %s", file))
	}

	fp = m[0].Code
	fp, err = uncompressFP(fp)
	if err != nil {
		return
	}
	return
}

func getFingerPrint(file string) (fp string, err error) {
	return getRangeFingerPrint(file, -1, -1)
}

func uncompressFP(fp string) (_fp string, err error) {
	var c []byte
	fp = strings.Replace(fp, "-", "+", -1)
	fp = strings.Replace(fp, "_", "/", -1)
	c, err = base64.StdEncoding.DecodeString(fp)
	if err != nil {
		return
	}

	byt := bytes.NewReader(c)
	var r io.ReadCloser
	r, err = zlib.NewReader(byt)
	if err != nil {
		return
	}

	defer r.Close()
	c, err = ioutil.ReadAll(r)

	if err != nil {
		return
	}

	if len(c)%10 != 0 {
		err = errors.New("length doesn't match")
		return
	}

	half := len(c) / 2
	var _a, _b int64
	var a, b []byte
	var result []string
	for i := 0; i < half; i += 5 {
		a = c[i : i+5]
		b = c[half+i : half+i+5]

		_a, err = strconv.ParseInt(string(a), 16, 0)
		if err != nil {
			return
		}
		_b, err = strconv.ParseInt(string(b), 16, 0)
		if err != nil {
			return
		}
		result = append(result, strconv.Itoa(int(_b)), strconv.Itoa(int(_a)))
	}
	_fp = strings.Join(result, " ")
	return
}
