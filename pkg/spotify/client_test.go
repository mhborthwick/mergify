package spotify

import (
	"errors"
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

	t.Run("failed request", func(t *testing.T) {
		mockClient := &http.Client{
			Transport: &mockRoundTripper{
				roundTripFunc: func(req *http.Request) (*http.Response, error) {
					return nil, errors.New("bad request")
				},
			},
		}
		s := Spotify{
			Client: mockClient,
			Token:  "mockToken",
		}
		userID, err := s.GetUserID()
		assert.Error(t, err, "expected failed")
		assert.Empty(t, userID, "userID should be empty")
	})

	t.Run("failed unmarshal", func(t *testing.T) {
		mockClient := &http.Client{
			Transport: &mockRoundTripper{
				roundTripFunc: func(req *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(strings.NewReader(`0`)),
					}, nil
				},
			},
		}
		s := Spotify{
			Client: mockClient,
			Token:  "mockToken",
		}
		userID, err := s.GetUserID()
		assert.Error(t, err, "expected unmarshal to fail")
		assert.Empty(t, userID, "userID should be empty")
	})
}
