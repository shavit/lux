package xiaohongshu

import (
	"testing"

	"github.com/iawia002/lux/extractors"
	"github.com/iawia002/lux/test"
)

func TestRumble(t *testing.T) {
	tests := []struct {
		name string
		args test.Args
	}{
		{
			name: "Note page",
			args: test.Args{
				URL:   "https://www.xiaohongshu.com/explore/640ef63e000000002702ad57",
                Title: "小窝日常｜住着会让人真切感受到幸福的卧室～",

			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			New().Extract(tt.args.URL, extractors.Options{})
		})
	}
}
