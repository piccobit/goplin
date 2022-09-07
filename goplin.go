// Package goplin provides an interface to the Data API of Joplin.

package goplin

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/imroc/req/v3"
)

type Client struct {
	handle   *req.Client
	port     int
	apiToken string
}

type Tag struct {
	ID       string `json:"id"`
	ParentID string `json:"parent_id"`
	Title    string `json:"title"`
}

const (
	joplinMinPortNum   = 41184
	joplinMaxPortNum   = 41194
	retriesGetApiToken = 20
)

var result struct {
	Items   []Tag `json:"items"`
	HasMore bool  `json:"has_more"`
}

func New(apiToken string) (*Client, error) {
	var retErr error

	joplinPortFound := false

	// In production, create a client explicitly and reuse it to send all requests
	// Use C() to create a client and set with chainable client settings.
	client := req.C().
		SetUserAgent("goplin").
		SetTimeout(5 * time.Second).
		DevMode()

	newClient := Client{
		handle:   client,
		port:     0,
		apiToken: apiToken,
	}

	for i := joplinMinPortNum; i <= joplinMaxPortNum; i++ {
		// Use R() to create a request and set with chainable request settings.
		resp, err := client.R(). // Use R() to create a request and set with chainable request settings.
						EnableDump(). // Enable dump at request level to help troubleshoot, log content only when an unexpected exception occurs.
						Get(fmt.Sprintf("http://localhost:%d/ping", i))
		if err != nil {
			retErr = err
			continue
		}

		if resp.IsError() {
			retErr = err
			continue
		}

		if resp.IsSuccess() {
			newClient.port = i

			if len(apiToken) == 0 {
				authToken, err := newClient.getAuthToken()
				if err != nil {
					retErr = err
					break
				}

				newClient.apiToken, err = newClient.getApiToken(authToken)
				if err != nil {
					retErr = err
					break
				}
			}

			joplinPortFound = true

			break
		}
	}

	if !joplinPortFound {
		return nil, retErr
	}

	return &newClient, nil
}

func (c *Client) getAuthToken() (string, error) {
	var token string

	var result struct {
		AuthToken string `json:"auth_token"`
	}

	resp, err := c.handle.R().
		SetResult(&result).
		Post(fmt.Sprintf("http://localhost:%d/auth", c.port))
	if err != nil {
		return token, err
	}

	if resp.IsError() {
		// handle response.
		err = fmt.Errorf("got error response, raw dump:\n%s", resp.Dump())

		return token, err
	}

	if resp.IsSuccess() {
		return result.AuthToken, nil
	}

	// handle response.
	err = fmt.Errorf("got unexpected response, raw dump:\n%s", resp.Dump())

	return token, err
}

func (c *Client) getApiToken(authToken string) (string, error) {
	var retErr error

	var result struct {
		Status   string `json:"status"`
		ApiToken string `json:"token,omitempty"`
	}

	retries := 0
	receivedApiToken := false

	for {
		resp, err := c.handle.R().
			SetQueryParam("auth_token", authToken).
			SetResult(&result).
			SetError(&result).
			Get(fmt.Sprintf("http://localhost:%d/auth/check", c.port))
		if err != nil {
			retErr = err
			break
		}

		if resp.IsError() {
			// Handle response.
			err = fmt.Errorf("got error response, raw dump:\n%s", resp.Dump())
			retErr = err

			break
		}

		if resp.IsSuccess() {
			if result.Status == "accepted" {
				receivedApiToken = true

				break
			} else if result.Status == "rejected" {
				err = errors.New("request rejected")
				retErr = err

				break
			} else if result.Status == "waiting" {
				retries++

				if retries < retriesGetApiToken {
					time.Sleep(time.Second)

					continue
				}

				retErr = fmt.Errorf("could not get an answer from user")

				break
			}
		}
	}

	if receivedApiToken {
		return result.ApiToken, nil
	}

	return "", retErr
}

func (c *Client) GetTags() ([]Tag, error) {
	var tags []Tag

	page := 1

	for {
		resp, err := c.handle.R().
			SetQueryParam("token", c.apiToken).
			SetQueryParam("page", strconv.Itoa(page)).
			SetResult(&result).
			SetError(&result).
			Get(fmt.Sprintf("http://localhost:%d/tags", c.port))
		if err != nil {
			return tags, err
		}

		if resp.IsError() {
			// handle response.
			err = fmt.Errorf("got error response, raw dump:\n%s", resp.Dump())

			return tags, err
		}

		if resp.IsSuccess() {
			for _, tag := range result.Items {
				tags = append(tags, tag)
			}

			if result.HasMore {
				page++

				continue
			} else {
				return tags, nil
			}
		}

		// handle response.
		err = fmt.Errorf("got unexpected response, raw dump:\n%s", resp.Dump())

		return tags, err
	}
}

func (c *Client) GetApiToken() string {
	return c.apiToken
}
