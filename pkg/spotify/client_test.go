package spotify

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockRoundTripper struct {
	roundTripFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.roundTripFunc(req)
}

func TestGetUserID(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		mockClient := &http.Client{
			Transport: &mockRoundTripper{
				roundTripFunc: func(req *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(strings.NewReader(`{"id": "user123"}`)),
					}, nil
				},
			},
		}
		s := Spotify{
			Client: mockClient,
			Token:  "mockToken",
		}
		userID, err := s.GetUserID()
		assert.NoError(t, err, "failed to unmarshal profile")
		assert.Equal(t, "user123", userID, "unexpected userID returned")
	})
}

func TestGetPlaylists(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		mockClient := &http.Client{
			Transport: &mockRoundTripper{
				roundTripFunc: func(req *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(strings.NewReader(`{"items": [{"id": "123", "name": "foo"}, {"id": "456", "name": "bar"}], "next": null}`)),
					}, nil
				},
			},
		}
		s := Spotify{
			Client: mockClient,
			Token:  "mockToken",
		}
		playlists, err := s.getPlaylists("user")
		expected := []Playlist{
			{ID: "123", Name: "foo"},
			{ID: "456", Name: "bar"},
		}
		assert.NoError(t, err, "failed to unmarshal profile")
		assert.Equal(t, expected, playlists, "unexpected playlists returned")
	})

	t.Run("pagination path", func(t *testing.T) {
		mockClient := &http.Client{
			Transport: &mockRoundTripper{
				roundTripFunc: func(req *http.Request) (*http.Response, error) {
					if req.URL.String() == "https://api.spotify.com/v1/users/user/playlists" {
						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       io.NopCloser(strings.NewReader(`{"items": [{"id": "123", "name": "foo"}, {"id": "456", "name": "bar"}], "next": "https://api.spotify.com/v1/users/user/playlists?offset=20"}`)),
						}, nil
					}
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(strings.NewReader(`{"items": [{"id": "789", "name": "baz"}], "next": null}`)),
					}, nil
				},
			},
		}
		s := Spotify{
			Client: mockClient,
			Token:  "mockToken",
		}
		playlists, err := s.getPlaylists("user")
		assert.NoError(t, err, "failed to unmarshal playlists")
		expected := []Playlist{
			{ID: "123", Name: "foo"},
			{ID: "456", Name: "bar"},
			{ID: "789", Name: "baz"},
		}
		assert.Equal(t, expected, playlists, "unexpected playlists")
	})
}
