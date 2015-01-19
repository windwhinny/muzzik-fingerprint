package muzzikfp

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Muzzik-Dev-Group/muzzik-fingerprint/http-client"
	"github.com/Muzzik-Dev-Group/muzzik-fingerprint/xiami"
	"io"
	"io/ioutil"
	"net/http"
	"os"
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

// Start 开始运行并分发 goroutine
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

// Next 获取下一个未处理的 Id 值, 如果所有 Id 都处理完成，则会触发结束
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

// Wait 等待所有的 goroutine 结束
func (set *FPWorkerSet) Wait() {
	<-set.Done
}

// HandleError 用来将错误返回给主进程。因为虾米有些 Id 并不包含歌曲，所以如果错误为 "empty response" 则略过。
func (set *FPWorkerSet) HandleError(err error) {
	if err.Error() != "empty response" {
		set.Errors <- err
	}
}

// CatchError 用来打印错误
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

// updateSolr 用来将歌曲信息更新至 solr
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

// Start 让 wf 开始工作，获取歌曲并计算其指纹
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

// Clean 删除歌曲文件
func (wf *FPWorkFlow) Clean() (err error) {
	if wf.Filename != "" {
		err = os.Remove(wf.Filename)
	}

	return
}

// Save 保存歌曲信息
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

// Remove 从 solr 里清除对应的歌曲信息
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

// GetMusic 通过 Id 获取歌曲文件和信息
func (wf *FPWorkFlow) GetMusic(id xiami.Id) (err error) {
	var name, fp string
	var music *xiami.Music
	music, name, err = download(id)
	if err != nil {
		return err
	}
	wf.Filename = name
	file := &AudioFile{Path: name}
	fp, err = file.getFingerPrint(false)
	if err != nil {
		return err
	}

	music.FingerPrint = fp
	music.XiamiId = id
	music.Id = strconv.Itoa(int(id))
	wf.Music = music
	return
}

// makeStorageFile 通过歌曲 Id 来建立文件并返回。例如 Id 为 12345的话，就返回 <MusicStorageDir>/123/45.m
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

// download 下载歌曲，并返回歌曲信息和保存的路劲
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
