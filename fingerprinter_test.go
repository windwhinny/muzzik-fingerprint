package muzzikfp

import (
	"github.com/Muzzik-Dev-Group/muzzik-fingerprint/xiami"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("FingerPrint", func() {
	Describe("Match", func() {
		var scanner *Scanner
		var wf *FPWorkFlow
		var music *xiami.Music
		var err error

		BeforeEach(func() {
			scanner = &Scanner{}
			wf = &FPWorkFlow{}

			err = wf.GetMusic(300)
			Expect(err).To(BeNil())
			err = wf.Save()
			Expect(err).To(BeNil())

			music = wf.Music
			scanner.Filename = wf.Filename
		})

		It("should successed", func() {
			err = scanner.Match()
			Expect(err).To(BeNil())
			Expect(scanner.Music).ToNot(BeNil())
			Expect(scanner.Music.Title).To(Equal(music.Title))
			Expect(scanner.Music.Artist).To(Equal(music.Artist))
		})
	})
})
