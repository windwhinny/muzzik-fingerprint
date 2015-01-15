package muzzikfp

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/windwhinny/muzzik-fingerprint/xiami"
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

	Describe("getFingerPrint", func() {
		var fp, file string
		var err error
		BeforeEach(func() {
			file = "./resource/music.mp3"
			fp, err = getFingerPrint(file)
			Expect(err).To(BeNil())
		})

		It("should get fp", func() {
			Expect(fp).ToNot(Equal(""))
		})
	})

	Describe("uncompressFP", func() {
		var code = "eJyl0UuOBSEIBdAtIR-B5QDK_pfQNbO78uIb9OQk3ggqAgAE3Kgr6waNG7xvwBXCG19q4caXW11nhXbjP--9dwa5camlmQKJm-sTYy_2AZz5kZoVMbE3bZdhhGtVSeZGrcluQ8HHfvYshLUs-1DIT_w3OwgKkEnMRKIlHTb7kP6c_MoOlKY0dIMYJnm1b2YJFGtPe74YgWUoqgasqAI8VMXv5QtOzUlIW1xaq3HrbGbEQpFQVnLEUHPwWmnS-yA9pF7ZYVktFUxXFTabnDQPMlD9lR2evpIjAiZm9DN_RzvYKuhXdvgBAzr-Kw=="
		var expect = "224080 10 732748 10 732748 10 732748 10 732748 10 732748 10 126281 12 66747 12 66747 12 66747 12 66747 12 66747 12 116426 13 680702 13 256337 13 537309 13 837051 13 927686 13 301079 49 37356 49 699680 49 907455 49 907455 49 907455 49 795818 78 907455 78 907455 78 907455 78 907455 78 907455 78 337155 14 547435 14 144341 14 1026159 14 1026159 14 1026159 14 758214 50 1026159 50 1026159 50 1026159 50 1026159 50 1026159 50 243827 14 97797 14 535353 14 850404 14 285221 14 588216 14 12320 48 282994 48 489997 48 707586 48 707586 48 707586 48 838146 78 707586 78 707586 78 707586 78 707586 78 707586 78 309174 10 205797 10 614268 10 994934 10 1000482 10 795994 10 476985 40 141944 40 592333 40 755198 40 755198 40 755198 40 389468 78 755198 78 755198 78 755198 78 755198 78 755198 78 888023 14 338839 14 480392 14 412470 14 412470 14 412470 14 332409 40 412470 40 412470 40 412470 40 412470 40 412470 40 89179 5 109062 5 178940 5 809256 5 809256 5 809256 5 580623 40 809256 40 809256 40 809256 40 809256 40 809256 40"
		It("should uncompress string", func() {
			s, err := uncompressFP(code)
			Expect(err).To(BeNil())
			Expect(s).To(Equal(expect))
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
