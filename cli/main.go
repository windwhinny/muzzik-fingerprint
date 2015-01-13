package main

import (
	"fmt"
	"github.com/jessevdk/go-flags"
	"github.com/windwhinny/muzzik-fingerprint"
	"os"
)

var options struct {
	Build bool   `short:"b" long:"build" description:"build database"`
	Host  string `short:"h" long:"host" description:"solr host address"`
}

func main() {
	args, err := flags.ParseArgs(&options, os.Args)

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	if options.Host != "" {
		muzzikfp.SolrHost = options.Host
	}

	if options.Build {
		set := &muzzikfp.FPWorkerSet{}
		set.MaxRoutine = 20
		set.MaxId = 1000000
		set.Start()
	} else {
		args = args[1:]
		if len(args) == 0 {
			fmt.Println("missing filename")
			os.Exit(1)
		}

		scanner := &muzzikfp.Scanner{}
		scanner.Filename = args[0]
		err = scanner.Match()
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		music := scanner.Music
		fmt.Printf("Artist: %s\nTitle: %s\n", music.Artist, music.Title)
	}
}
