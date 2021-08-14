package provider

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	"golang.org/x/oauth2"
)

type Client struct {
	baseURL *url.URL
	client  *http.Client
}

func NewClient(baseURL *url.URL) *Client {
	return &Client{
		baseURL: baseURL,
		client:  &http.Client{},
	}
}

func (c Client) Do(ctx context.Context, req *http.Request, response interface{}) error {
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		if err == io.EOF {
			return nil
		}
		return err
	}

	return nil
}

func (c Client) CreateRequest(tokenSrc oauth2.TokenSource, method, path string, body interface{}) (*http.Request, error) {
	url, err := c.baseURL.Parse(path)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, url.String(), nil)
	if err != nil {
		return nil, err
	}
	token, err := tokenSrc.Token()
	if err != nil {
		return nil, err
	}
	token.SetAuthHeader(req)

	return req, err
}
