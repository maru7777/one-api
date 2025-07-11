package replicate

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/Laisky/errors/v2"
	"github.com/gin-gonic/gin"
	"golang.org/x/image/webp"
	"golang.org/x/sync/errgroup"

	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/relay/adaptor/openai"
	"github.com/songquanpeng/one-api/relay/meta"
	"github.com/songquanpeng/one-api/relay/model"
)

var errNextLoop = errors.New("next_loop")

// ImageHandler handles the response from the image creation or remix request
func ImageHandler(c *gin.Context, resp *http.Response) (
	*model.ErrorWithStatusCode, *model.Usage) {
	if resp.StatusCode != http.StatusCreated {
		payload, _ := io.ReadAll(resp.Body)
		return openai.ErrorWrapper(
				errors.Errorf("bad_status_code [%d]%s", resp.StatusCode, string(payload)),
				"bad_status_code", http.StatusInternalServerError),
			nil
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return openai.ErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError), nil
	}

	respData := new(ImageResponse)
	if err = json.Unmarshal(respBody, respData); err != nil {
		return openai.ErrorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError), nil
	}

	for {
		err = func() error {
			// get task
			taskReq, err := http.NewRequestWithContext(c.Request.Context(),
				http.MethodGet, respData.URLs.Get, nil)
			if err != nil {
				return errors.Wrap(err, "new request")
			}

			taskReq.Header.Set("Authorization", "Bearer "+meta.GetByContext(c).APIKey)
			taskResp, err := http.DefaultClient.Do(taskReq)
			if err != nil {
				return errors.Wrap(err, "get task")
			}
			defer taskResp.Body.Close()

			if taskResp.StatusCode != http.StatusOK {
				payload, _ := io.ReadAll(taskResp.Body)
				return errors.Errorf("bad status code [%d]%s",
					taskResp.StatusCode, string(payload))
			}

			taskBody, err := io.ReadAll(taskResp.Body)
			if err != nil {
				return errors.Wrap(err, "read task response")
			}

			taskData := new(ImageResponse)
			if err = json.Unmarshal(taskBody, taskData); err != nil {
				return errors.Wrap(err, "decode task response")
			}

			switch taskData.Status {
			case "succeeded":
			case "failed", "canceled":
				return errors.Errorf("task failed, [%s]%s", taskData.Status, taskData.Error)
			default:
				time.Sleep(time.Second * 3)
				return errNextLoop
			}

			output, err := taskData.GetOutput()
			if err != nil {
				return errors.Wrap(err, "get output")
			}
			if len(output) == 0 {
				return errors.New("response output is empty")
			}

			var mu sync.Mutex
			var pool errgroup.Group
			respBody := &openai.ImageResponse{
				Created: taskData.CompletedAt.Unix(),
				Data:    []openai.ImageData{},
			}

			for _, imgOut := range output {
				imgOut := imgOut
				pool.Go(func() error {
					// download image
					downloadReq, err := http.NewRequestWithContext(c.Request.Context(),
						http.MethodGet, imgOut, nil)
					if err != nil {
						return errors.Wrap(err, "new request")
					}

					imgResp, err := http.DefaultClient.Do(downloadReq)
					if err != nil {
						return errors.Wrap(err, "download image")
					}
					defer imgResp.Body.Close()

					if imgResp.StatusCode != http.StatusOK {
						payload, _ := io.ReadAll(imgResp.Body)
						return errors.Errorf("bad status code [%d]%s",
							imgResp.StatusCode, string(payload))
					}

					imgData, err := io.ReadAll(imgResp.Body)
					if err != nil {
						return errors.Wrap(err, "read image")
					}

					imgData, err = ConvertImageToPNG(imgData)
					if err != nil {
						return errors.Wrap(err, "convert image")
					}

					mu.Lock()
					respBody.Data = append(respBody.Data, openai.ImageData{
						B64Json: fmt.Sprintf("data:image/png;base64,%s",
							base64.StdEncoding.EncodeToString(imgData)),
					})
					mu.Unlock()

					return nil
				})
			}

			if err := pool.Wait(); err != nil {
				if len(respBody.Data) == 0 {
					return errors.WithStack(err)
				}

				logger.Error(c, fmt.Sprintf("some images failed to download: %+v", err))
			}

			c.JSON(http.StatusOK, respBody)
			return nil
		}()
		if err != nil {
			if errors.Is(err, errNextLoop) {
				continue
			}

			return openai.ErrorWrapper(err, "image_task_failed", http.StatusInternalServerError), nil
		}

		break
	}

	return nil, nil
}

// ConvertImageToPNG converts a WebP image to PNG format
func ConvertImageToPNG(webpData []byte) ([]byte, error) {
	// bypass if it's already a PNG image
	if bytes.HasPrefix(webpData, []byte("\x89PNG")) {
		return webpData, nil
	}

	// check if is jpeg, convert to png
	if bytes.HasPrefix(webpData, []byte("\xff\xd8\xff")) {
		img, _, err := image.Decode(bytes.NewReader(webpData))
		if err != nil {
			return nil, errors.Wrap(err, "decode jpeg")
		}

		var pngBuffer bytes.Buffer
		if err := png.Encode(&pngBuffer, img); err != nil {
			return nil, errors.Wrap(err, "encode png")
		}

		return pngBuffer.Bytes(), nil
	}

	// Decode the WebP image
	img, err := webp.Decode(bytes.NewReader(webpData))
	if err != nil {
		return nil, errors.Wrap(err, "decode webp")
	}

	// Encode the image as PNG
	var pngBuffer bytes.Buffer
	if err := png.Encode(&pngBuffer, img); err != nil {
		return nil, errors.Wrap(err, "encode png")
	}

	return pngBuffer.Bytes(), nil
}
