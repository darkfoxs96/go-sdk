package apivideosdk

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
)

type UploadFileReader struct {
	file        io.Reader
	maxReadBuf  int
	readBuf     []byte
	readBufSize int

	writeBuf       []byte
	maxWriteBuf    int
	writeBufSize   int
	filePartWriter io.Writer
	formWriter     *multipart.Writer
}

func (r *UploadFileReader) Init(file io.Reader, filePartWriter io.Writer, formWriter *multipart.Writer) () {
	r.filePartWriter = filePartWriter
	r.file = file
	r.formWriter = formWriter
}

func (r *UploadFileReader) Write(p []byte) (n int, err error) {
	n = len(p)
	if n > r.maxWriteBuf {
		n = r.maxWriteBuf
	}

	for i := 0; i < n; i++ {
		r.writeBuf[i] = p[i]
	}

	r.writeBufSize = n
	return
}

func (r *UploadFileReader) Read(p []byte) (n int, err error) {
	if r.writeBufSize != 0 {
		outSize := r.writeBufSize
		for i := 0; i < outSize; i++ {
			p[i] = r.writeBuf[i]
			r.writeBufSize--
		}
		return outSize, nil
	}

	r.readBufSize, err = r.file.Read(r.readBuf)
	if err == io.EOF && r.readBufSize == 0 {
		err = io.EOF
	} else {
		err = nil
	}

	if err == io.EOF {
		_ = r.formWriter.Close()
	} else {
		_, _ = r.filePartWriter.Write(r.readBuf[:r.readBufSize])
	}

	outSize := r.writeBufSize
	for i := 0; i < outSize; i++ {
		p[i] = r.writeBuf[i]
	}
	r.readBufSize = 0
	r.writeBufSize = 0
	return outSize, err
}

func (c *Client) prepareRangeFromRequest(urlStr string, file io.Reader, fileHeaders *multipart.FileHeader) ([]*http.Request, error) {
	requests := []*http.Request{}

	body := new(UploadFileReader)
	body.maxReadBuf = 256
	body.maxWriteBuf = 512
	body.readBuf = make([]byte, body.maxReadBuf, body.maxReadBuf)
	body.writeBuf = make([]byte, body.maxWriteBuf, body.maxWriteBuf)
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", fileHeaders.Filename)
	if err != nil {
		return nil, err
	}

	body.Init(file, part, writer)

	req, err := c.prepareRequest(http.MethodPost, urlStr, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	requests = append(requests, req)
	return requests, nil
}

func (s *VideosService) UploadFromRequest(videoID string, file io.Reader, fileHeaders *multipart.FileHeader) (*Video, error) {
	path := fmt.Sprintf("%s/%s/source", videosBasePath, videoID)

	requests, err := s.client.prepareRangeFromRequest(path, file, fileHeaders)

	if err != nil {
		return nil, err
	}

	v := new(Video)

	for _, req := range requests {
		_, err = s.client.do(req, v)

		if err != nil {
			return nil, err
		}
	}
	return v, nil
}
