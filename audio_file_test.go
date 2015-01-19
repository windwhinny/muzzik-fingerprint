package muzzikfp

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type MockAudioFile struct {
	Duration int
	Range    []int
}

func (file *MockAudioFile) getDuration() (int, error) {
	return file.Duration, nil
}

func (file *MockAudioFile) getRangeFingerPrint(start int, end int, compress bool) (string, error) {
	file.Range = append(file.Range, start, end)
	return "", nil
}

var _ = Describe("AudioFile", func() {

	Describe("getFingerPrint", func() {
		var fp string
		var file *AudioFile
		var err error
		BeforeEach(func() {
			file = &AudioFile{Path: "./resource/music.mp3"}

			fp, err = file.getFingerPrint(true)
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
			s, err := UncompressFP(code)
			Expect(err).To(BeNil())
			Expect(s).To(Equal(expect))
		})
	})

	Describe("formatDurationStr", func() {
		It("should get duration from string", func() {
			duration, err := formatDurationStr([]byte("01:02:03"))
			Expect(err).To(BeNil())
			Expect(duration).To(Equal(1*60*60 + 2*60 + 3))
		})
	})

	Describe("GetFPs", func() {
		var file *MockAudioFile

		BeforeEach(func() {
			file = &MockAudioFile{}
			file.Duration = 60
		})

		It("should return 3 part of fp", func() {
			fps, err := GetFPs(file)
			Expect(err).To(BeNil())
			Expect(len(fps)).To(Equal(3))
			r := []int{0, 10, 25, 35, 50, 60}

			for k, v := range file.Range {
				Expect(v).To(Equal(r[k]))
			}
		})
	})
})
