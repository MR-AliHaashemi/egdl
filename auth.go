package egdl

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"time"
)

type EGError struct {
	Code    string `json:"errorCode"`
	Message string `json:"errorMessage"`
}

func findEGError(buf io.Reader) error {
	data := &EGError{}
	if err := json.NewDecoder(buf).Decode(data); err != nil {
		return err
	}
	return data
}

func (err *EGError) Error() string {
	return err.Message
}

type GrantedToken struct {
	AccessToken    string    `json:"access_token"`
	ExpiresIn      int       `json:"expires_in"`
	ExpiresAt      time.Time `json:"expires_at"`
	TokenType      string    `json:"token_type"`
	ClientID       string    `json:"client_id"`
	InternalClient bool      `json:"internal_client"`
	ClientService  string    `json:"client_service"`
}

func oAuthClientCredentials() (*GrantedToken, error) {
	requestPayloads := url.Values{}
	requestPayloads.Set("grant_type", "client_credentials")
	requestPayloads.Set("token_type", "eg1")

	req, err := http.NewRequest(
		http.MethodPost,
		"https://account-public-service-prod.ol.epicgames.com/account/api/oauth/token",
		bytes.NewReader([]byte(requestPayloads.Encode())),
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "UELauncher/14.1.6-21413796+++Portal+Release-Live Windows/10.0.22622.1.256.64bit")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "basic MzRhMDJjZjhmNDQxNGUyOWIxNTkyMTg3NmRhMzZmOWE6ZGFhZmJjY2M3Mzc3NDUwMzlkZmZlNTNkOTRmYzc2Y2Y=")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, findEGError(req.Body)
	}

	data := &GrantedToken{}
	err = json.NewDecoder(resp.Body).Decode(data)
	return data, err
}
