package muzzikfp

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("FingerPrint", func() {
	Describe("Matcher", func() {
		Describe("Match", func() {
			var matcher *Matcher

			MockSolrQuery := func(fp string) (musics Musics, err error) {
				music1 := &Music{}
				music1.Id = "1"
				music1.Score = 0.1
				music2 := &Music{}
				music2.Id = "2"
				music2.Score = 0.3
				music3 := &Music{}
				music3.Id = "3"
				music3.Score = 0.6
				switch fp {
				case "1":
					musics = Musics{music1, music2, music3}
				case "2":
					musics = Musics{music1, music2}
				case "3":
					musics = Musics{music1, music2, music3}
				}
				return
			}

			BeforeEach(func() {
				matcher = &Matcher{}
				matcher.FPs = []string{"1", "2", "3"}
			})

			It("should return best match", func() {
				matcher.querySolr = MockSolrQuery
				music, err := matcher.Match()
				Expect(err).To(BeNil())
				Expect(music.Score).To(Equal(float32(0.4)))
				Expect(music.Id).To(Equal("3"))
			})
		})
	})
})
