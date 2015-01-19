package main

import (
	"encoding/json"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net/http"
)

type MockResponse struct {
	StatusCode int
	Data       []byte
	header     http.Header
}

func (res *MockResponse) Header() http.Header {
	return res.header
}

func (res *MockResponse) Write(data []byte) (writed int, err error) {
	res.Data = append(res.Data, data...)
	writed = len(data)
	return
}

func (res *MockResponse) WriteHeader(statusCode int) {
	res.StatusCode = statusCode
}

var _ = Describe("Main", func() {
	Describe("HttpHandler", func() {
		var req *http.Request
		var res *MockResponse

		BeforeEach(func() {
			req = &http.Request{}
			req.Header = make(http.Header)
			res = &MockResponse{}
			res.header = make(http.Header)
		})

		It("should only accpet POST method", func() {
			req.Method = "GET"
			queryHandler(res, req)
			Expect(res.StatusCode).To(Equal(http.StatusMethodNotAllowed))
			httpErr := &HttpError{}
			json.Unmarshal(res.Data, httpErr)
			Expect(httpErr.Message).To(Equal("only POST allowed"))
		})

	})
})
