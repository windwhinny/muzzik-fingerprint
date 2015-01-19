package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Muzzik-Dev-Group/muzzik-fingerprint"
	"github.com/jessevdk/go-flags"
	"io"
	"net/http"
	"os"
	"runtime/debug"
)

var options struct {
	Host string `short:"h" long:"host" description:"query host address, default to localhost:8090"`
}

func handleErr(err error) {
	if err != nil {
		fmt.Println(err.Error())
		debug.PrintStack()
		os.Exit(1)
	}
}

func main() {
	args, err := flags.ParseArgs(&options, os.Args)
	handleErr(err)

	if len(args) < 2 {
		err = errors.New("select a file")
		handleErr(err)
	}

	var fps []string
	var buf bytes.Buffer
	var url string
	var res *http.Response
	var req *http.Request
	music := &muzzikfp.Music{}

	if options.Host != "" {
		url = "http://" + options.Host + "/query"
	} else {
		url = "http://localhost:8090/query"
	}

	file := &muzzikfp.AudioFile{}
	file.Path = args[1]
	fps, err = muzzikfp.GetFPs(file)
	handleErr(err)
	err = json.NewEncoder(&buf).Encode(fps)
	handleErr(err)
	req, err = http.NewRequest("POST", url, &buf)
	handleErr(err)
	res, err = http.DefaultClient.Do(req)
	buf.Reset()
	defer res.Body.Close()
	_, err = io.Copy(&buf, res.Body)
	handleErr(err)
	err = json.NewDecoder(&buf).Decode(music)
	handleErr(err)
	fmt.Printf("Title:%s\nArtist:%s\nScore:%f\n", music.Title, music.Artist, music.Score)
	return
}
