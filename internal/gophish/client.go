package gophish

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func NewClient(baseURL, apiKey string, insecure bool) *Client {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: insecure}, //nolint:gosec
	}
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		apiKey:     apiKey,
		httpClient: &http.Client{Transport: transport},
	}
}

func (c *Client) do(method, path string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, c.baseURL+path, bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", c.apiKey)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return c.httpClient.Do(req)
}

func (c *Client) GetCampaigns() ([]Campaign, error) {
	resp, err := c.do(http.MethodGet, "/api/campaigns/", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, apiError(resp)
	}

	var campaigns []Campaign
	return campaigns, json.NewDecoder(resp.Body).Decode(&campaigns)
}

func (c *Client) GetCampaignByName(name string) (*Campaign, error) {
	campaigns, err := c.GetCampaigns()
	if err != nil {
		return nil, err
	}
	for i := range campaigns {
		if campaigns[i].Name == name {
			return &campaigns[i], nil
		}
	}
	return nil, fmt.Errorf("campaign %q not found", name)
}

func (c *Client) CreateCampaign(req CreateCampaignRequest) (*Campaign, error) {
	resp, err := c.do(http.MethodPost, "/api/campaigns/", req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, apiError(resp)
	}

	var campaign Campaign
	return &campaign, json.NewDecoder(resp.Body).Decode(&campaign)
}

func (c *Client) GetCampaignResults(id int64) (*CampaignResults, error) {
	resp, err := c.do(http.MethodGet, fmt.Sprintf("/api/campaigns/%d/results", id), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, apiError(resp)
	}

	var results CampaignResults
	return &results, json.NewDecoder(resp.Body).Decode(&results)
}

func (c *Client) DeleteCampaign(id int64) error {
	resp, err := c.do(http.MethodDelete, fmt.Sprintf("/api/campaigns/%d", id), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return apiError(resp)
	}
	return nil
}

func (c *Client) GetTemplates() ([]Template, error) {
	resp, err := c.do(http.MethodGet, "/api/templates/", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, apiError(resp)
	}

	var templates []Template
	return templates, json.NewDecoder(resp.Body).Decode(&templates)
}

func (c *Client) GetTemplateByName(name string) (*Template, error) {
	templates, err := c.GetTemplates()
	if err != nil {
		return nil, err
	}
	for i := range templates {
		if templates[i].Name == name {
			return &templates[i], nil
		}
	}
	return nil, fmt.Errorf("template %q not found", name)
}

func (c *Client) CreateTemplate(t Template) (*Template, error) {
	resp, err := c.do(http.MethodPost, "/api/templates/", t)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, apiError(resp)
	}

	var created Template
	return &created, json.NewDecoder(resp.Body).Decode(&created)
}

func (c *Client) UpdateTemplate(t Template) (*Template, error) {
	resp, err := c.do(http.MethodPut, fmt.Sprintf("/api/templates/%d", t.ID), t)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, apiError(resp)
	}

	var updated Template
	return &updated, json.NewDecoder(resp.Body).Decode(&updated)
}

func (c *Client) DeleteTemplate(id int64) error {
	resp, err := c.do(http.MethodDelete, fmt.Sprintf("/api/templates/%d", id), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return apiError(resp)
	}
	return nil
}

func (c *Client) GetGroups() ([]Group, error) {
	resp, err := c.do(http.MethodGet, "/api/groups/", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, apiError(resp)
	}

	var groups []Group
	return groups, json.NewDecoder(resp.Body).Decode(&groups)
}

func (c *Client) GetGroupByName(name string) (*Group, error) {
	groups, err := c.GetGroups()
	if err != nil {
		return nil, err
	}
	for i := range groups {
		if groups[i].Name == name {
			return &groups[i], nil
		}
	}
	return nil, fmt.Errorf("group %q not found", name)
}

func (c *Client) CreateGroup(name string, targets []Target) (*Group, error) {
	if targets == nil {
		targets = []Target{}
	}
	body := struct {
		Name    string   `json:"name"`
		Targets []Target `json:"targets"`
	}{Name: name, Targets: targets}

	resp, err := c.do(http.MethodPost, "/api/groups/", body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, apiError(resp)
	}

	var group Group
	return &group, json.NewDecoder(resp.Body).Decode(&group)
}

func (c *Client) UpdateGroup(group Group) (*Group, error) {
	resp, err := c.do(http.MethodPut, fmt.Sprintf("/api/groups/%d", group.ID), group)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, apiError(resp)
	}

	var updated Group
	return &updated, json.NewDecoder(resp.Body).Decode(&updated)
}

// ImportGroupCSV uploads a CSV file to GoPhish's bulk-import endpoint and
// returns the targets GoPhish parsed from it. It does not create or modify
// any group — the caller is responsible for that.
func (c *Client) ImportGroupCSV(path string) ([]Target, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	part, err := mw.CreateFormFile("file", filepath.Base(path))
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(part, f); err != nil {
		return nil, err
	}
	if err := mw.Close(); err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/api/import/group", &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", c.apiKey)
	req.Header.Set("Content-Type", mw.FormDataContentType())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, apiError(resp)
	}

	var targets []Target
	return targets, json.NewDecoder(resp.Body).Decode(&targets)
}

func (c *Client) DeleteGroup(id int64) error {
	resp, err := c.do(http.MethodDelete, fmt.Sprintf("/api/groups/%d", id), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return apiError(resp)
	}
	return nil
}

func (c *Client) GetSMTPProfiles() ([]SMTP, error) {
	resp, err := c.do(http.MethodGet, "/api/smtp/", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, apiError(resp)
	}

	var profiles []SMTP
	return profiles, json.NewDecoder(resp.Body).Decode(&profiles)
}

func (c *Client) CreateSMTPProfile(s SMTP) (*SMTP, error) {
	resp, err := c.do(http.MethodPost, "/api/smtp/", s)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, apiError(resp)
	}

	var created SMTP
	return &created, json.NewDecoder(resp.Body).Decode(&created)
}

func (c *Client) DeleteSMTPProfile(id int64) error {
	resp, err := c.do(http.MethodDelete, fmt.Sprintf("/api/smtp/%d", id), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return apiError(resp)
	}
	return nil
}

func apiError(resp *http.Response) error {
	b, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("API error %s: %s", resp.Status, strings.TrimSpace(string(b)))
}
