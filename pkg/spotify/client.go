package spotify

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Spotify struct {
	Token  string
	Client *http.Client
	UserID string
}

type Profile struct {
	ID string `json:"id"`
}

const API = "https://api.spotify.com/v1"

func (s *Spotify) handleRequest(method, endpoint string, body io.Reader) ([]byte, error) {
	if s.Client == nil {
		s.Client = &http.Client{}
	}
	url := API + endpoint
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", s.Token))
	resp, err := s.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status code %d", resp.StatusCode)
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	return respBody, nil
}

func (s *Spotify) getProfile() (*Profile, error) {
	body, err := s.handleRequest("GET", "/me", nil)
	if err != nil {
		return nil, err
	}
	var profile Profile
	if err := json.Unmarshal(body, &profile); err != nil {
		return nil, fmt.Errorf("failed to unmarshal profile: %w", err)
	}
	return &profile, nil
}

func (s *Spotify) GetUserID() (string, error) {
	profile, err := s.getProfile()
	if err != nil {
		return "", err
	}
	return profile.ID, nil
}
