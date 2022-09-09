// Package goplin provides an interface to the Data API of Joplin.

package goplin

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
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
	Type     int    `json:"type_,omitempty"`
}

type Note struct {
	ID       string `json:"id"`
	ParentID string `json:"parent_id"`
	Title    string `json:"title"`
	Type     int    `json:"type_,omitempty"`
}

type tagsResult struct {
	Items   []Tag `json:"items"`
	HasMore bool  `json:"has_more"`
}

type notesResult struct {
	Notes   []Note `json:"items"`
	HasMore bool   `json:"has_more"`
}

const (
	joplinMinPortNum   = 41184
	joplinMaxPortNum   = 41194
	retriesGetApiToken = 20
)

func New(apiToken string) (*Client, error) {
	var retErr error

	joplinPortFound := false

	// In production, create a client explicitly and reuse it to send all requests
	// Use C() to create a client and set with chainable client settings.
	client := req.C().
		SetUserAgent("goplin").
		SetTimeout(5 * time.Second)

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

func (c *Client) GetTag(id string) (Tag, error) {
	var tag Tag

	resp, err := c.handle.R().
		SetPathParam("id", id).
		SetQueryParam("token", c.apiToken).
		SetResult(&tag).
		SetError(&tag).
		Get(fmt.Sprintf("http://localhost:%d/tags/{id}", c.port))
	if err != nil {
		return tag, err
	}

	if resp.IsError() {
		if resp.StatusCode == 404 {
			err = fmt.Errorf("could not find tag with IDs '%s", id)

		} else {
			err = fmt.Errorf("got error response, raw dump:\n%s", resp.Dump())
		}

		return tag, err
	}

	if resp.IsSuccess() {
		return tag, nil
	}

	// handle response.
	err = fmt.Errorf("got unexpected response, raw dump:\n%s", resp.Dump())

	return tag, err
}

func (c *Client) GetNote(id string) (Note, error) {
	var note Note

	resp, err := c.handle.R().
		SetPathParam("id", id).
		SetQueryParam("token", c.apiToken).
		SetResult(&note).
		SetError(&note).
		Get(fmt.Sprintf("http://localhost:%d/notes/{id}", c.port))
	if err != nil {
		return note, err
	}

	if resp.IsError() {
		if resp.StatusCode == 404 {
			err = fmt.Errorf("could not find note with ID '%s", id)
		} else {
			err = fmt.Errorf("got error response, raw dump:\n%s", resp.Dump())
		}

		return note, err
	}

	if resp.IsSuccess() {
		return note, nil
	}

	// Handle response.
	err = fmt.Errorf("got unexpected response, raw dump:\n%s", resp.Dump())

	return note, err
}

func (c *Client) GetNotesByTag(id string, orderBy string, orderDir string) ([]Note, error) {
	var result notesResult
	var notes []Note

	page := 1

	queryParams := map[string]string{
		"token":  c.apiToken,
		"fields": "id,parent_id,title",
		"page":   strconv.Itoa(page),
	}

	if len(orderBy) != 0 {
		queryParams["order_by"] = orderBy
	}

	if len(orderDir) != 0 {
		queryParams["order_dir"] = strings.ToUpper(orderDir)
	}

	for {
		resp, err := c.handle.R().
			SetPathParam("id", id).
			SetQueryParams(queryParams).
			SetResult(&result).
			SetError(&result).
			Get(fmt.Sprintf("http://localhost:%d/tags/{id}/notes", c.port))
		if err != nil {
			return notes, err
		}

		if resp.IsError() {
			if resp.StatusCode == 404 {
				err = fmt.Errorf("could not find note with IDs '%s", id)
			} else {
				err = fmt.Errorf("got error response, raw dump:\n%s", resp.Dump())
			}

			return notes, err
		}

		if resp.IsSuccess() {
			for _, note := range result.Notes {
				notes = append(notes, note)
			}

			if result.HasMore {
				page++

				queryParams["page"] = strconv.Itoa(page)

				continue
			} else {
				return notes, nil
			}
		}

		// handle response.
		err = fmt.Errorf("got unexpected response, raw dump:\n%s", resp.Dump())

		return notes, err
	}
}

func (c *Client) GetAllNotes(orderBy string, orderDir string) ([]Note, error) {
	var result notesResult
	var notes []Note

	page := 1

	queryParams := map[string]string{
		"token":  c.apiToken,
		"fields": "id,parent_id,title",
		"page":   strconv.Itoa(page),
	}

	if len(orderBy) != 0 {
		queryParams["order_by"] = orderBy
	}

	if len(orderDir) != 0 {
		queryParams["order_dir"] = strings.ToUpper(orderDir)
	}

	for {
		resp, err := c.handle.R().
			SetQueryParams(queryParams).
			SetResult(&result).
			SetError(&result).
			Get(fmt.Sprintf("http://localhost:%d/notes", c.port))
		if err != nil {
			return notes, err
		}

		if resp.IsError() {
			// handle response.
			err = fmt.Errorf("got error response, raw dump:\n%s", resp.Dump())

			return notes, err
		}

		if resp.IsSuccess() {
			for _, note := range result.Notes {
				notes = append(notes, note)
			}

			if result.HasMore {
				page++

				queryParams["page"] = strconv.Itoa(page)

				continue
			} else {
				return notes, nil
			}
		}

		// handle response.
		err = fmt.Errorf("got unexpected response, raw dump:\n%s", resp.Dump())

		return notes, err
	}
}

func (c *Client) GetAllTags(orderBy string, orderDir string) ([]Tag, error) {
	var result tagsResult
	var tags []Tag

	page := 1

	queryParams := map[string]string{
		"token":  c.apiToken,
		"fields": "id,parent_id,title",
		"page":   strconv.Itoa(page),
	}

	if len(orderBy) != 0 {
		queryParams["order_by"] = orderBy
	}

	if len(orderDir) != 0 {
		queryParams["order_dir"] = strings.ToUpper(orderDir)
	}

	for {
		resp, err := c.handle.R().
			SetQueryParams(queryParams).
			SetResult(&result).
			SetError(&result).
			Get(fmt.Sprintf("http://localhost:%d/tags/", c.port))
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

				queryParams["page"] = strconv.Itoa(page)

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

func (c *Client) DeleteTag(id string) error {
	resp, err := c.handle.R().
		SetPathParam("id", id).
		SetQueryParam("token", c.apiToken).
		Delete(fmt.Sprintf("http://localhost:%d/tags/{id}", c.port))
	if err != nil {
		return err
	}

	if resp.IsError() {
		// handle response.
		err = fmt.Errorf("got error response, raw dump:\n%s", resp.Dump())

		return err
	}

	if resp.IsSuccess() {
		return nil
	}

	// handle response.
	err = fmt.Errorf("got unexpected response, raw dump:\n%s", resp.Dump())

	return err
}

func (c *Client) DeleteTagFromNote(tagID string, noteID string) error {
	resp, err := c.handle.R().
		SetPathParam("tagID", tagID).
		SetPathParam("noteID", noteID).
		SetQueryParam("token", c.apiToken).
		Delete(fmt.Sprintf("http://localhost:%d/tags/{tagID}/notes/{noteID}", c.port))
	if err != nil {
		return err
	}

	if resp.IsError() {
		// handle response.
		err = fmt.Errorf("got error response, raw dump:\n%s", resp.Dump())

		return err
	}

	if resp.IsSuccess() {
		return nil
	}

	// handle response.
	err = fmt.Errorf("got unexpected response, raw dump:\n%s", resp.Dump())

	return err
}

func (c *Client) GetApiToken() string {
	return c.apiToken
}
