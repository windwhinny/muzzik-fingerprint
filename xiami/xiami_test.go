package xiami

import (
	"github.com/cheggaaa/pb"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Xiami", func() {

	Describe("GetMusic", func() {
		var music *Music
		var err error
		BeforeEach(func() {
			music, err = GetMusic(100)
			Expect(err).To(BeNil())
		})

		It("should return music", func() {
			Expect(music).ToNot(BeNil())
			Expect(music.Artist).ToNot(Equal(""))
			Expect(music.Title).ToNot(Equal(""))
		})

		It("should not be blocked", func() {
			progress := pb.StartNew(100)
			for i := 1; i <= 100; i++ {
				music, err := GetMusic(Id(i))
				if err != nil && err.Error() == "empty response" {
				} else {
					Expect(err).To(BeNil())
					Expect(music.Artist).ToNot(Equal(""))
				}
				progress.Increment()
			}
			progress.FinishPrint("Done")
		})
	})

	Describe("Track", func() {
		var track *XiamiTrack
		BeforeEach(func() {
			track = &XiamiTrack{}
			track.Url = `8h2fmFF1mtDe53195utFii112ph2dbd-3Elt%l.%%83_2a1a14-lp2ec222%kd8f944%%F.oFF43e81f62%53mxm12_Fy2f96%5EA5i%%_la%e1345E-%.a222.u3818dE%n`
		})

		Describe("DecodeUrl", func() {

			BeforeEach(func() {
				err := track.DecodeUrl()
				Expect(err).To(BeNil())
			})

			It("should decode url", func() {
				Expect(track.Url).To(Equal(`http://m5.file.xiami.com/1/1/1/2_212824_l.mp3?auth_key=22d82e8eda81f115b1ff9383da9664d1-1420934400-0-null`))
			})
		})
	})

})
