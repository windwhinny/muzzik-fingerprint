package muzzikfp

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/windwhinny/muzzik-fingerprint/http-client"
	"github.com/windwhinny/muzzik-fingerprint/xiami"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type FPWorkerSet struct {
	workers    []*FPWorkFlow
	number     uint64
	Done       chan bool
	Errors     chan error
	mutex      *sync.Mutex
	MaxRoutine int
	StartId    int
	EndId      int
	SaveToSolr bool
}

type FPWorkFlow struct {
	Music      *xiami.Music
	Filename   string
	SaveToSolr bool
}

var SolrHost string
var MusicStorageDir string

type JSON map[string]interface{}

func (set *FPWorkerSet) Start() {
	set.Errors = make(chan error, 100)
	set.Done = make(chan bool)
	set.mutex = &sync.Mutex{}
	set.number = uint64(set.StartId)
	set.CatchError()
	for i := 0; i < set.MaxRoutine; i++ {
		wf := &FPWorkFlow{}
		wf.SaveToSolr = set.SaveToSolr
		go func() {
			for {
				number := set.Next()
				if number == 0 {
					break
				}

				err := wf.Start(xiami.Id(number))

				if err != nil {
					set.HandleError(err)
				}
			}
		}()

		set.workers = append(set.workers, wf)
		time.Sleep(500 * time.Millisecond)
	}

	set.Wait()
}

func (set *FPWorkerSet) Next() (number uint64) {
	number = atomic.LoadUint64(&set.number)
	fmt.Printf("%s, loaded %d\n", time.Now().Format(time.RFC3339), set.number)
	atomic.AddUint64(&set.number, 1)
	number = atomic.LoadUint64(&set.number)

	if number > uint64(set.EndId) {
		number = 0
		set.Done <- true
	}
	return
}

func (set *FPWorkerSet) Wait() {
	<-set.Done
}

func (set *FPWorkerSet) HandleError(err error) {
	if err.Error() != "empty response" {
		set.Errors <- err
	}
}

func (set *FPWorkerSet) CatchError() {
	go func() {
		for {
			err := <-set.Errors
			if err != nil {
				fmt.Println(err.Error())
			}
		}
	}()
}

func updateSolr(data JSON) (err error) {
	var buf bytes.Buffer
	err = json.NewEncoder(&buf).Encode(data)
	if err != nil {
		return
	}

	var res *http.Response
	var link string
	var linkTmp = "http://%s/solr/fp/update?commitWithin=600000"
	if SolrHost == "" {
		link = fmt.Sprintf(linkTmp, "localhost:8080")
	} else {
		link = fmt.Sprintf(linkTmp, SolrHost)
	}
	res, err = http.Post(link, "Content-type:application/json", &buf)
	if err != nil {
		return
	}
	defer res.Body.Close()
	if res.StatusCode > 400 {
		var buf []byte
		buf, err = ioutil.ReadAll(res.Body)
		if err != nil {
			return
		}
		err = errors.New(fmt.Sprintf("solr return %d\n%s\n", res.StatusCode, string(buf)))
	}

	return
}

func (wf *FPWorkFlow) Start(id xiami.Id) (err error) {
	err = wf.GetMusic(id)
	if err != nil {
		return
	}

	if wf.SaveToSolr {
		err = wf.Save()
		if err != nil {
			return
		}
	}

	return
}

func (wf *FPWorkFlow) Clean() (err error) {
	if wf.Filename != "" {
		err = os.Remove(wf.Filename)
	}

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

func (wf *FPWorkFlow) GetMusic(id xiami.Id) (err error) {
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

func makeStorageFile(id xiami.Id) (file *os.File, err error) {
	path := strconv.Itoa(int(id))
	var strs []string
	for i := 0; i < len(path); i += 3 {
		var end int
		if i+3 > len(path) {
			end = len(path)
		} else {
			end = i + 3
		}
		strs = append(strs, path[i:end])
	}
	strs[len(strs)-1] = strs[len(strs)-1] + ".m"
	path = strings.Join(strs, "/")
	dir := strings.Join(strs[:len(strs)-1], "/")

	if dir != "" {
		if MusicStorageDir == "" {
			dir = "/tmp/music/" + dir
		} else {
			dir = MusicStorageDir + "/" + dir
		}

		err = os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			return
		}
	}
	if MusicStorageDir == "" {
		path = "/tmp/music/" + path
	} else {
		path = MusicStorageDir + "/" + path
	}
	file, err = os.Create(path)
	return
}

func download(id xiami.Id) (music *xiami.Music, name string, err error) {
	var res *http.Response
	music, err = xiami.GetMusic(id)
	if err != nil {
		return
	}
	res, err = httpClient.Get(music.Url)

	if err != nil {
		return
	}

	var file *os.File
	file, err = makeStorageFile(id)
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
