package handler_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/frohwerk/deputy-backend/cmd/imgmatch/handler"
	"github.com/frohwerk/deputy-backend/cmd/imgmatch/handler/mocks"
	"github.com/frohwerk/deputy-backend/internal/database"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	done := make(chan interface{})

	matcher := mocks.NewMockMatcher(ctrl)
	matcher.EXPECT().
		Match(gomock.Eq("172.30.1.1:5000/example/image:1.1")).
		Return([]database.File{{Id: "1", Name: "app-1.0.tar.gz", Digest: "sha256:abcb", Parent: ""}}, nil).
		Times(1)

	linker := mocks.NewMockImageLinker(ctrl)
	linker.EXPECT().
		AddLink(gomock.Eq("172.30.1.1:5000/example/image:1.1"), gomock.Eq("1")).
		DoAndReturn(func(image, file string) (*database.ImageLink, error) {
			defer func() { done <- nil }()
			return &database.ImageLink{Id: image, FileId: file}, nil
		}).
		Times(1)

	server := httptest.NewServer(handler.New(matcher, linker))
	defer server.Close()

	params := url.Values{}
	params.Set("image", "172.30.1.1:5000/example/image:1.1")
	client := server.Client()
	resp, err := client.PostForm(server.URL, params)

	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusAccepted, resp.StatusCode, "expecting http status 202 (accepted)")
		t.Log(time.Now(), "http request complete")
		<-done
		t.Log(time.Now(), "asynchronous processing complete")
	}
}
