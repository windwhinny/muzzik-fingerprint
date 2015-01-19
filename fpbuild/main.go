package main

import (
	"errors"
	"fmt"
	"github.com/Muzzik-Dev-Group/muzzik-fingerprint"
	"github.com/jessevdk/go-flags"
	"os"
	"runtime/debug"
	"strconv"
)

var options struct {
	MusicDir   string `short:"d" long:"dir" description:"music storage directory"`
	Host       string `short:"h" long:"host" description:"solr host address, default to localhost:8080"`
	Routine    int    `short:"r" long:"routines" description:"routines to run"`
	SaveToSolr bool   `short:"s" long:"saveToSolr" description:"save to solr searchengine"`
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
	args = args[1:]
	if options.Host != "" {
		muzzikfp.SolrHost = options.Host
	}

	if options.MusicDir != "" {
		var stat os.FileInfo
		dir, err := os.Open(options.MusicDir)
		handleErr(err)
		stat, err = dir.Stat()
		handleErr(err)
		if !stat.IsDir() {
			err = errors.New(fmt.Sprintf("%s is not directory", dir.Name()))
			handleErr(err)
		}

		muzzikfp.MusicStorageDir = options.MusicDir
	}

	set := &muzzikfp.FPWorkerSet{}
	if options.Routine == 0 {
		set.MaxRoutine = 20
	} else {
		set.MaxRoutine = options.Routine
	}
	var start, end int64
	if len(args) == 2 {
		start, err = strconv.ParseInt(args[0], 10, 0)
		handleErr(err)
		end, err = strconv.ParseInt(args[1], 10, 0)
		handleErr(err)
	} else if len(args) == 1 {
		start, err = strconv.ParseInt(args[0], 10, 0)
		handleErr(err)
		end = 100000
	} else {
		start = 0
		end = 100000
	}

	set.SaveToSolr = options.SaveToSolr
	set.EndId = int(end)
	set.StartId = int(start)
	set.Start()
}
