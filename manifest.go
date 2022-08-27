// ToDo: Refactor the whole file...
package fndl

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"

	"github.com/MR-AliHaashemi/egs"
	"github.com/MR-AliHaashemi/egs/egerrors"
	"github.com/er-azh/egmanifest"
)

const (
	LauncherURL       = "https://launcher-public-service-prod06.ol.epicgames.com/launcher/api/public/assets/v2/platform/"
	FortniteNameSpace = "fn"
	FortniteID        = "4fe75bbc5a674f4f9b356b5c90567da5"
	FortniteName      = "Fortnite"
	FortniteLabel     = "Live"
)

type Platform string

const (
	Windows Platform = "Windows"
	Android Platform = "Android"
)

type ManifestInfo struct {
	AppName      string `json:"appName"`
	LabelName    string `json:"labelName"`
	BuildVersion string `json:"buildVersion"`
	Hash         string `json:"hash"`
	Manifests    []struct {
		URI         string `json:"uri"`
		QueryParams []struct {
			Name  string `json:"name"`
			Value string `json:"value"`
		} `json:"queryParams"`
	} `json:"manifests"`
}

type ManifestProvider struct {
	auth *egs.GrantedToken
}

func NewManifestProvider() (*ManifestProvider, error) {
	auth, err := egs.OAuthClientCredentials("Windows/10.0.22000.1.256.64bit", egs.ClientLauncherApp2)
	if err != nil {
		return nil, err
	}

	return &ManifestProvider{auth: auth}, nil
}

func (m *ManifestProvider) GetManifestInfo(platform Platform) (*ManifestInfo, error) {
	finalURL := LauncherURL + string(platform) + "/namespace/" + FortniteNameSpace + "/catalogItem/" + FortniteID + "/app/" + FortniteName + "/label/" + FortniteLabel
	req, err := http.NewRequest(http.MethodGet, finalURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "UELauncher/13.1.2-18458102+++Portal+Release-Live Windows/10.0.22000.1.256.64bit")
	req.Header.Set("Authorization", "bearer "+m.auth.AccessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, egerrors.FindError(resp.Body)
	}

	data := &struct {
		Elements []ManifestInfo `json:"elements"`
	}{}

	if err = json.NewDecoder(resp.Body).Decode(data); err != nil {
		return nil, err
	} else if len(data.Elements) == 0 {
		return nil, errors.New("zero elements found")
	}
	return &data.Elements[0], err
}

func (m *ManifestProvider) Download(data *ManifestInfo) (*egmanifest.BinaryManifest, error) {
	manifest := data.Manifests[len(data.Manifests)-1]

	newURL, err := url.Parse(manifest.URI)
	if err != nil {
		return nil, err
	}

	query := newURL.Query()
	for _, param := range manifest.QueryParams {
		query.Add(param.Name, param.Value)
	}
	newURL.RawQuery = query.Encode()

	return DownloadManifest(newURL.String())
}

func DownloadManifest(manifestURL string) (*egmanifest.BinaryManifest, error) {
	resp, err := http.Get(manifestURL)
	if err != nil || resp.StatusCode != http.StatusOK {
		return nil, err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return egmanifest.ParseManifest(bytes.NewReader(raw))
}

func LoadManifest(f io.ReadSeeker) (*egmanifest.BinaryManifest, error) {
	return egmanifest.ParseManifest(f)
}
