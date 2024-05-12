package urlShortenerClient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	urlShortener "github.com/Sleeps17/linker/internal/clients/url-shortener"
	"io"
	"net/http"
)

type Client struct {
	url      string
	username string
	password string
}

func New(host, port string, username, password string) urlShortener.UrlShortener {
	return &Client{
		url:      fmt.Sprintf("http://%s:%s/", host, port),
		username: username,
		password: password,
	}
}

func (c *Client) SaveURL(ctx context.Context, Url, alias string) (string, error) {
	type (
		Request struct {
			Url   string `json:"url"`
			Alias string `json:"alias"`
		}

		Response struct {
			Status string `json:"status"`
			Error  string `json:"error"`
			Alias  string `json:"alias"`
		}
	)

	body, err := json.Marshal(Request{Url: Url, Alias: alias})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, c.url, bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}

	req.SetBasicAuth(c.username, c.password)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	defer func() { _ = resp.Body.Close() }()

	var response Response
	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return "", err
	}

	if response.Status != "OK" {
		return "", errors.New(response.Error)
	}

	return response.Alias, nil
}

func (c *Client) DeleteURL(ctx context.Context, alias string) error {
	type (
		Request struct {
			Alias string `json:"alias"`
		}

		Response struct {
			Status string `json:"status"`
			Error  string `json:"error"`
		}
	)

	body, err := json.Marshal(Request{Alias: alias})
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodDelete, c.url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	req.SetBasicAuth(c.username, c.password)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer func() { _ = resp.Body.Close() }()

	jsonResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var response Response
	if err := json.Unmarshal(jsonResp, &response); err != nil {
		return err
	}

	if response.Status != "OK" {
		return errors.New(response.Error)
	}

	return nil
}
