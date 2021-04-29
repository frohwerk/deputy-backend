package envs_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/frohwerk/deputy-backend/cmd/server/envs"
	"github.com/frohwerk/deputy-backend/internal/database"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type creatorMock struct {
	names []string
}

func (s *creatorMock) Create(name string) (*database.Env, error) {
	s.names = append(s.names, name)
	return &database.Env{Id: uuid.NewString(), Name: name}, nil
}

func TestMain(t *testing.T) {
	store := new(creatorMock)
	server := httptest.NewServer(http.HandlerFunc(envs.Create(store)))
	defer server.Close()

	t.Log(server.URL)

	client := server.Client()
	url := fmt.Sprintf("%s", server.URL)
	resp, err := client.Post(url, "application/json", strings.NewReader(`{"name": "test"}`))

	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
		t.Log("Location:", resp.Header.Get("Location"))
		// assert.Equal(t, "", resp.Header.Get("Location"))
	}
}
