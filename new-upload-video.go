package apivideosdk

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
)

func (c *Client) prepareRangeFromRequest(urlStr string, file multipart.File, fileHeaders *multipart.FileHeader) ([]*http.Request, error) {
	var bufSize int64
	if fileHeaders.Size > c.chunkSize && c.chunkSize != 0 {
		bufSize = c.chunkSize
	} else {
		bufSize = fileHeaders.Size
	}

	buf := make([]byte, bufSize)
	requests := []*http.Request{}
	startByte := 0
	for {
		bytesread, err := file.Read(buf)
		if err != nil {
			if err != io.EOF {
				return nil, err
			}
			break
		}

		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)
		part, err := writer.CreateFormFile("file", fileHeaders.Filename)
		if err != nil {
			return nil, err
		}
		part.Write(buf)

		err = writer.Close()
		if err != nil {
			return nil, err
		}

		req, err := c.prepareRequest(http.MethodPost, urlStr, body)
		if err != nil {
			return nil, err
		}

		req.Header.Set("Content-Type", writer.FormDataContentType())

		if fileHeaders.Size > c.chunkSize && c.chunkSize != 0 {
			ranges := fmt.Sprintf("bytes %d-%d/%d", startByte, (startByte+bytesread)-1, fileHeaders.Size)
			req.Header.Set("Content-Range", ranges)
			startByte = startByte + bytesread
		}

		if err != nil {
			return nil, err
		}

		requests = append(requests, req)
	}
	return requests, nil
}

func (s *VideosService) UploadFromRequest(videoID string, file multipart.File, fileHeaders *multipart.FileHeader) (*Video, error) {
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
