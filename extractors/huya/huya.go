package huya

import (
	"github.com/pkg/errors"

	"github.com/iawia002/lux/extractors"
	"github.com/iawia002/lux/request"
	"github.com/iawia002/lux/utils"
)

func init() {
	extractors.Register("huya", New())
}

type extractor struct{}

const huyaVideoHost = "https://videotx-platform.cdn.huya.com/"

// New returns a huya extractor.
func New() extractors.Extractor {
	return &extractor{}
}

func (e *extractor) Extract(url string, option extractors.Options) ([]*extractors.Data, error) {
	headers := map[string]string{}
	headers["Referer"] = "https://www.huya.com"
	headers["User-Agent"] = "Mozilla/5.0 (Linux; Android 8.0; Pixel 2 Build/OPD3.170816.012) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.88 Mobile Safari/537.36 Edg/87.0.664.66"
	html, err := request.Get(url, url, headers)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var title string
	titleDesc := utils.MatchOneOf(html, `<h1>(.+?)</h1>`)
	if len(titleDesc) > 1 {
		title = titleDesc[1]
	} else {
		title = "huya video"
	}

	var videoUrl string
	videoDesc := utils.MatchOneOf(html, `src="\/\/.+huya.com\/(.*)" poster=(.+?)`)
	if len(videoDesc) > 1 {
		videoUrl = huyaVideoHost + videoDesc[1]
	} else {
		return nil, errors.WithStack(extractors.ErrURLParseFailed)
	}

	size, err := request.Size(videoUrl, url)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	urlData := &extractors.Part{
		URL:  videoUrl,
		Size: size,
		Ext:  "mp4",
	}
	quality := "normal"
	streams := map[string]*extractors.Stream{
		quality: {
			Parts:   []*extractors.Part{urlData},
			Size:    size,
			Quality: quality,
		},
	}
	return []*extractors.Data{
		{
			Site:    "虎牙 huya.com",
			Title:   title,
			Type:    extractors.DataTypeVideo,
			Streams: streams,
			URL:     url,
		},
	}, nil
}
