package xiaohongshu

import (
	"compress/flate"
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
	"regexp"
    "strings"

	"github.com/pkg/errors"

	"github.com/iawia002/lux/extractors"
	"github.com/iawia002/lux/request"
)

func init() {
	extractors.Register("xiaohongshu", New())
}

type extractor struct{}

func New() extractors.Extractor {
	return &extractor{}
}

// pagePayload is the initial state when the page loads.
// Use this to read metadata about the note page.
type pagePayload struct {
    Note note `json:"note"`
}

// note object has pagination and the note object inside
type note struct {
    Note struct {
        Description string `json:"desc"`
        ImageList json.RawMessage `json:"imageList"`
        NoteID string `json:"noteId"`
        Title string `json:"title"`
        Type string `json:"type"` // ex. video
        User struct {
            UserID string `json:"userId"`
            Nickname string `json:"nickname"`
            Avatar string `json:"avatar"`
        } `json:"user"`
        Video struct {
            Media struct {
                VideoID int64 `json:"videoId"`
                Video struct {
                    MD5 string `json:"md5"`
                } `json:"video"`
                Stream struct {
                    H265 []*noteVideoStream `json:"h265"`
                    AV1 []*noteVideoStream `json:"av1"`
                    H264 []*noteVideoStream `json:"h264"`
                } `json:"stream"`
            } `json:"media"`
        } `json:"video"`
    } `json:"note"`
}

type noteVideoStream struct {
    AudioBitrate uint16 `json:"audioBitrate"`
    AudioChannels uint8 `json:"audioChannels"`
    AudioCodec string `json:"audioCodec"`
    BackupURLs []string `json:"backupUrls"`
    Duration uint64 `json:"duration"`
    FPS uint8 `json:"fps"`
    Height uint16 `json:"height"`
    MasterURL string `json:"masterUrl"`
    QualityType string `json:"qualityType"`
    Size uint64 `json:"size"`
    VideoBitrate uint32 `json:"videoBitrate"`
    VideoCodec string `json:"videoCodec"`
    Width uint16 `json:"width"`
}

func (e extractor) Extract(url string, option extractors.Options) ([]*extractors.Data, error) {
	res, err := request.Request(http.MethodGet, url, nil, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	defer res.Body.Close() // nolint

	var reader io.ReadCloser
	switch res.Header.Get("Content-Encoding") {
	case "gzip":
		reader, _ = gzip.NewReader(res.Body)
	case "deflate":
		reader = flate.NewReader(res.Body)
	default:
		reader = res.Body
	}
	defer reader.Close() // nolint

	b, err := io.ReadAll(reader)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var resp *pagePayload
    if resp, err = parsePayload(b); err != nil {
		return nil, errors.WithStack(extractors.ErrURLParseFailed)
	}
    
    note := resp.Note.Note
    h264Vids := note.Video.Media.Stream.H264
    if len(h264Vids) == 0 {
		return nil, errors.WithStack(extractors.ErrURLParseFailed)
    }
	
    streams := make(map[string]*extractors.Stream, len(h264Vids))
    for _, vid := range h264Vids {
        var parts []*extractors.Part
        for _, backupURL := range vid.BackupURLs {
            part := &extractors.Part {
                URL: backupURL,
                Size: int64(vid.Size),
            }
            parts = append(parts, part)
        }
    
        streams[vid.QualityType] = &extractors.Stream{
		Parts:   parts,
		Size:    int64(vid.Size),
        Quality: vid.QualityType,
        }
    }
    
	return []*extractors.Data{
		{
			Site:    "Xiaohongshu xiaohongshu.com",
			Title:   note.Title,
			Type:    extractors.DataTypeVideo,
			Streams: streams,
			URL:     url,
		},
	}, nil
}

var rePayload = regexp.MustCompile(`\<script\>(?s).*window\.__INITIAL_STATE__\s=\s(.*)\<\/script\>`)

func parsePayload(b []byte) (*pagePayload, error) {
	matched := rePayload.FindSubmatch(b)
    if len(matched) < 1 {
		return nil, errors.WithStack(extractors.ErrURLParseFailed)
    }
    body := strings.ReplaceAll(string(matched[1]), "undefined", "null")
    
    var resp *pagePayload
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		return nil, errors.WithStack(extractors.ErrURLParseFailed)
    }
    
    return resp, nil
}
