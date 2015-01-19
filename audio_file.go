package muzzikfp

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

var durationRegexp *regexp.Regexp

func init() {
	var err error
	durationRegexp, err = regexp.Compile("[0-9]{2}:[0-9]{2}:[0-9]{2}")
	if err != nil {
		panic(err)
	}
}

type AudioFile struct {
	Path     string
	Duration int
}

type AudioFileParser interface {
	getDuration() (int, error)
	getRangeFingerPrint(int, int, bool) (string, error)
}

// GetFPs 取出文件前中后三段各10秒的音频指纹, 如果时长小于10秒则只取一段，小于30秒则取两段
func GetFPs(file AudioFileParser) (fps []string, err error) {
	var duration int
	duration, err = file.getDuration()

	if err != nil {
		return
	}
	var fp string
	var leftR, middleR, rightR []int
	leftR = append(leftR, 0)
	if duration <= 10 {
		leftR = append(leftR, duration)
	} else {
		leftR = append(leftR, 10)
	}

	fp, err = file.getRangeFingerPrint(leftR[0], leftR[1], true)
	if err != nil {
		return
	}
	fps = append(fps, fp)

	if duration >= 30 {
		middleR = append(middleR, (duration/2)-5, 10)
		fp, err = file.getRangeFingerPrint(middleR[0], middleR[1], true)
		if err != nil {
			return
		}
		fps = append(fps, fp)
	}

	if duration > 10 {
		rightR = append(rightR, duration-10, 10)
		fp, err = file.getRangeFingerPrint(rightR[0], rightR[1], true)
		if err != nil {
			return
		}
		fps = append(fps, fp)
	}
	return
}

// formatDurationStr 会将格式为 00:00:00 的字符串转换为对应的时间,单位为秒
func formatDurationStr(str []byte) (duration int, err error) {
	times := bytes.Split(str, []byte(":"))

	if len(times) != 3 {
		err = errors.New("unable to get duration")
		return
	}
	var hour, min, sec int

	hour, err = strconv.Atoi(string(times[0]))
	if err != nil {
		return
	}
	min, err = strconv.Atoi(string(times[1]))
	if err != nil {
		return
	}
	sec, err = strconv.Atoi(string(times[2]))
	if err != nil {
		return
	}

	duration = hour*60*60 + min*60 + sec
	return
}

// getDuration 获取传入音频文件的时长，如果文件不是音频文件，则返回错误
func (file *AudioFile) getDuration() (duration int, err error) {
	if file.Duration != 0 {
		duration = file.Duration
		return
	}

	var buf bytes.Buffer
	c1 := exec.Command("ffmpeg", "-i", file.Path, "2>&1")
	c2 := exec.Command("grep", "Duration")

	r, w := io.Pipe()

	c1.Stderr = w
	c1.Stdout = w
	c2.Stdin = r
	c2.Stdout = &buf

	c1.Start()
	c2.Start()
	c1.Wait()
	w.Close()
	c2.Wait()

	duraStr := durationRegexp.Find(buf.Bytes())
	if len(duraStr) == 0 {
		err = errors.New("unable to get duration")
		return
	}

	duration, err = formatDurationStr(duraStr)
	file.Duration = duration
	return
}

// getRangeFingerPrint 获取歌曲音频指纹。start 和 end 分别代表需要截取指纹的区域，如果都为 -1 则代表全部歌曲。
func (file *AudioFile) getRangeFingerPrint(start int, length int, commpress bool) (fp string, err error) {
	if file.Path == "" {
		err = errors.New("no file")
		return
	}
	var cmd *exec.Cmd
	if start < 0 || length < 0 {
		cmd = exec.Command("echoprint-codegen", file.Path)
	} else {
		cmd = exec.Command("echoprint-codegen", file.Path, strconv.Itoa(start), strconv.Itoa(length))
	}
	var buf bytes.Buffer
	cmd.Stdout = &buf

	err = cmd.Run()
	if err != nil {
		return
	}
	var m []*struct {
		Code string `json:"code"`
	}

	err = json.Unmarshal(buf.Bytes(), &m)
	if err != nil {
		return
	}

	if len(m) == 0 {
		err = errors.New(fmt.Sprintf("cant get fingerprint of file %s", file))
	}

	fp = m[0].Code
	if commpress == false {
		fp, err = UncompressFP(fp)
		if err != nil {
			return
		}
	}
	return
}

func (file *AudioFile) getFingerPrint(compress bool) (fp string, err error) {
	return file.getRangeFingerPrint(-1, -1, compress)
}

// UncompressFP 将压缩过的指纹转换为原始指纹
func UncompressFP(fp string) (_fp string, err error) {
	var c []byte
	fp = strings.Replace(fp, "-", "+", -1)
	fp = strings.Replace(fp, "_", "/", -1)
	c, err = base64.StdEncoding.DecodeString(fp)
	if err != nil {
		return
	}

	byt := bytes.NewReader(c)
	var r io.ReadCloser
	r, err = zlib.NewReader(byt)
	if err != nil {
		return
	}

	defer r.Close()
	c, err = ioutil.ReadAll(r)

	if err != nil {
		return
	}

	if len(c)%10 != 0 {
		err = errors.New("length doesn't match")
		return
	}

	half := len(c) / 2
	var _a, _b int64
	var a, b []byte
	var result []string
	for i := 0; i < half; i += 5 {
		a = c[i : i+5]
		b = c[half+i : half+i+5]

		_a, err = strconv.ParseInt(string(a), 16, 0)
		if err != nil {
			return
		}
		_b, err = strconv.ParseInt(string(b), 16, 0)
		if err != nil {
			return
		}
		result = append(result, strconv.Itoa(int(_b)), strconv.Itoa(int(_a)))
	}
	_fp = strings.Join(result, " ")
	return
}
