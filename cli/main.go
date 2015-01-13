package main

import (
	"github.com/windwhinny/muzzik-fingerprint"
)

func main() {
	set := &muzzikfp.FPWorkerSet{}
	set.MaxRoutine = 20
	set.MaxId = 1000000
	set.Start()
}
