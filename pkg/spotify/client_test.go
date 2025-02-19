package spotify

import (
	"encoding/json"
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

	t.Run("pagination", func(t *testing.T) {
		mockClient := &http.Client{
			Transport: &mockRoundTripper{
				roundTripFunc: func(req *http.Request) (*http.Response, error) {
					if req.URL.String() == "http://localhost:3000/users/user/playlists" {
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

func TestGetTracksFromPlaylist(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		mockClient := &http.Client{
			Transport: &mockRoundTripper{
				roundTripFunc: func(req *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(strings.NewReader(`{"items": [{"track": {"uri": "123"}}, {"track": {"uri": "456"}}], "next": null}`)),
					}, nil
				},
			},
		}
		s := Spotify{
			Client: mockClient,
			Token:  "mockToken",
		}
		tracks, err := s.getTracksFromPlaylist("mockPlaylistID")
		expected := []PlaylistTrack{
			{
				Track: Track{
					URI: "123",
				},
			},
			{
				Track: Track{
					URI: "456",
				},
			},
		}
		assert.NoError(t, err, "failed to unmarshal response")
		assert.Equal(t, expected, tracks, "unexpected tracks returned")
	})

	t.Run("pagination", func(t *testing.T) {
		mockClient := &http.Client{
			Transport: &mockRoundTripper{
				roundTripFunc: func(req *http.Request) (*http.Response, error) {
					if req.URL.String() == "http://localhost:3000/playlists/mockPlaylistID/tracks" {
						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       io.NopCloser(strings.NewReader(`{"items": [{"track": {"uri": "123"}}, {"track": {"uri": "456"}}], "next": "https://api.spotify.com/v1/users/user/playlists?offset=20"}`)),
						}, nil
					}
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(strings.NewReader(`{"items": [{"track": {"uri": "111"}}, {"track": {"uri": "222"}}], "next": null}`)),
					}, nil
				},
			},
		}
		s := Spotify{
			Client: mockClient,
			Token:  "mockToken",
		}
		tracks, err := s.getTracksFromPlaylist("mockPlaylistID")
		expected := []PlaylistTrack{
			{
				Track: Track{
					URI: "123",
				},
			},
			{
				Track: Track{
					URI: "456",
				},
			},
			{
				Track: Track{
					URI: "111",
				},
			},
			{
				Track: Track{
					URI: "222",
				},
			},
		}
		assert.NoError(t, err, "failed to unmarshal response")
		assert.Equal(t, expected, tracks, "unexpected tracks returned")
	})
}

func TestAddTracksToPlaylist(t *testing.T) {
	t.Run("tracks are batched correctly", func(t *testing.T) {
		var requests []map[string][]string
		mockClient := &http.Client{
			Transport: &mockRoundTripper{
				roundTripFunc: func(req *http.Request) (*http.Response, error) {
					var body map[string][]string
					if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
						t.Fatalf("failed to decode request body: %v", err)
					}
					requests = append(requests, body)
					mockResponse := `{"snapshot_id": "mockSnapshot123"}`
					return &http.Response{
						StatusCode: http.StatusCreated,
						Body:       io.NopCloser(strings.NewReader(mockResponse)),
					}, nil
				},
			},
		}
		s := Spotify{
			Client: mockClient,
			Token:  "mockToken",
		}
		trackIDs := []string{"track1", "track2", "track3", "track4", "track5"}
		batchSize := 2
		response, _ := s.AddTracksToPlaylist("mockPlaylistID", trackIDs, batchSize)
		assert.Equal(t, "mockSnapshot123", response)
		assert.Equal(t, 3, len(requests), "unexpected number of batches")
		expectedBatches := []map[string][]string{
			{"uris": {"track1", "track2"}},
			{"uris": {"track3", "track4"}},
			{"uris": {"track5"}},
		}
		assert.Equal(t, expectedBatches, requests, "unexpected batch content")
	})
}

func Test_GetPlaylistTracksIDs(t *testing.T) {
	t.Run("omits duplicate tracks", func(t *testing.T) {
		t.Run("happy path", func(t *testing.T) {
			mockClient := &http.Client{
				Transport: &mockRoundTripper{
					roundTripFunc: func(req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       io.NopCloser(strings.NewReader(`{"items": [{"track": {"uri": "123"}}, {"track": {"uri": "456"}}, {"track": {"uri": "456"}}], "next": null}`)),
						}, nil
					},
				},
			}
			s := Spotify{
				Client: mockClient,
				Token:  "mockToken",
			}
			tracks, err := s.GetPlaylistTrackIDs([]string{"playListId1", "playListId2"})
			expected := []string{"123", "456"}
			assert.NoError(t, err, "failed to unmarshal response")
			assert.Equal(t, expected, tracks, "unexpected tracks returned")
		})
	})
}
