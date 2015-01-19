package muzzikfp

import (
	"github.com/Muzzik-Dev-Group/muzzik-fingerprint/xiami"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"os"
)

var _ = Describe("CodeGen", func() {

	Describe("download", func() {
		var file string
		var music *xiami.Music
		var err error
		BeforeEach(func() {
			music, file, err = download(100)
			Expect(err).To(BeNil())
		})

		AfterEach(func() {
			err = os.Remove(file)
			Expect(err).To(BeNil())
		})

		It("should download file automatic", func() {
			Expect(file).ToNot(Equal(""))
			var f *os.File
			f, err = os.Open(file)
			Expect(err).To(BeNil())
			defer f.Close()
		})
	})

	Describe("makeStorageFile", func() {
		var file *os.File
		var err error
		BeforeEach(func() {
			file, err = makeStorageFile(12345678)
			Expect(err).To(BeNil())
		})

		AfterEach(func() {
			err = os.RemoveAll(file.Name())
			Expect(err).To(BeNil())
		})

		It("should return a file within directory", func() {
			stat, err := file.Stat()
			Expect(err).To(BeNil())
			Expect(stat.IsDir()).To(Equal(false))
			Expect(file.Name()).To(Equal("/tmp/music/123/456/78.m"))
		})
	})

	Describe("Save", func() {
		var wf *FPWorkFlow
		var err error

		AfterEach(func() {
			err = wf.Remove()
			Expect(err).To(BeNil())
		})

		It("should successed", func() {
			wf = &FPWorkFlow{}
			err = wf.GetMusic(200)
			Expect(err).To(BeNil())
			err = wf.Save()
			Expect(err).To(BeNil())
		})
	})
})
