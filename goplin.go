// Package goplin provides an interface to the Data API of Joplin.

package goplin

import (
	"encoding/json"
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
	ID                   string `json:"id"`
	ParentID             string `json:"parent_id"`
	Title                string `json:"title"`
	CreatedTime          int    `json:"created_time,omitempty"`
	UpdatedTime          int    `json:"updated_time,omitempty"`
	UserCreatedTime      int    `json:"user_created_time,omitempty"`
	UserUpdatedTime      int    `json:"user_updated_time,omitempty"`
	EncryptionCipherText string `json:"encryption_cipher_text,omitempty"`
	EncryptionApplied    int    `json:"encryption_applied,omitempty"`
	IsShared             int    `json:"is_shared,omitempty"`
	Type                 int    `json:"type_,omitempty"`
}

type Note struct {
	ID                   string  `json:"id"`
	ParentID             string  `json:"parent_id"`
	Title                string  `json:"title"`
	Body                 string  `json:"body,omitempty"`
	CreatedTime          int     `json:"created_time,omitempty"`
	UpdatedTime          int     `json:"updated_time,omitempty"`
	IsConflict           int     `json:"is_conflict,omitempty"`
	Latitude             float64 `json:"latitude,omitempty"`
	Longitude            float64 `json:"longitude,omitempty"`
	Altitude             float64 `json:"altitude,omitempty"`
	Author               string  `json:"author,omitempty"`
	SourceURL            string  `json:"source_url,omitempty"`
	IsTodo               int     `json:"is_todo,omitempty"`
	TodoDue              int     `json:"todo_due,omitempty"`
	TodoCompleted        int     `json:"todo_completed,omitempty"`
	Source               string  `json:"source,omitempty"`
	SourceApplication    string  `json:"source_application,omitempty"`
	ApplicationData      string  `json:"application_data,omitempty"`
	Order                float64 `json:"order,omitempty"`
	UserCreatedTime      int     `json:"user_created_time,omitempty"`
	UserUpdatedTime      int     `json:"user_updated_time,omitempty"`
	EncryptionCipherText string  `json:"encryption_cipher_text,omitempty"`
	EncryptionApplied    int     `json:"encryption_applied,omitempty"`
	MarkupLanguage       int     `json:"markup_language,omitempty"`
	IsShared             int     `json:"is_shared,omitempty"`
	ShareID              string  `json:"share_id,omitempty"`
	ConflictOriginalID   string  `json:"conflict_original_id,omitempty"`
	MasterKeyID          string  `json:"master_key_id,omitempty"`
	BodyHTML             string  `json:"body_html,omitempty"`
	BaseURL              string  `json:"base_url,omitempty"`
	ImageDataURL         string  `json:"image_data_url,omitempty"`
	CropRect             string  `json:"crop_rect,omitempty"`
	Type                 int     `json:"type_,omitempty"`
}

type Notebook struct {
	ID                      string `json:"id"`
	ParentID                string `json:"parent_id"`
	Title                   string `json:"title"`
	CreatedTime             int    `json:"created_time,omitempty"`
	UpdatedTime             int    `json:"updated_time,omitempty"`
	UserCreatedTime         int    `json:"user_created_time,omitempty"`
	UserUpdatedTime         int    `json:"user_updated_time,omitempty"`
	EncryptionCipherText    string `json:"encryption_cipher_text,omitempty"`
	EncryptionApplied       int    `json:"encryption_applied,omitempty"`
	EncryptionBlobEncrypted int    `json:"encryption_blob_encrypted,omitempty"`
	IsShared                int    `json:"is_shared,omitempty"`
	ShareID                 string `json:"share_id,omitempty"`
	MasterKeyID             string `json:"master_key_id,omitempty"`
	Icon                    string `json:"icon,omitempty"`
}

type Resource struct {
	ID                      string `json:"id"`
	ParentID                string `json:"parent_id"`
	Title                   string `json:"title"`
	Mime                    string `json:"mime,omitempty"`
	Filename                string `json:"filename,omitempty"`
	CreatedTime             int    `json:"created_time,omitempty"`
	UpdatedTime             int    `json:"updated_time,omitempty"`
	FileExtension           string `json:"file_extension,omitempty"`
	EncryptionCipherText    string `json:"encryption_cipher_text,omitempty"`
	EncryptionApplied       int    `json:"encryption_applied,omitempty"`
	EncryptionBlobEncrypted int    `json:"encryption_blob_encrypted,omitempty"`
	Size                    int    `json:"size,omitempty"`
	IsShared                int    `json:"is_shared,omitempty"`
	ShareID                 string `json:"share_id,omitempty"`
	MasterKeyID             string `json:"master_key_id,omitempty"`
}

type Event struct {
	ID               string `json:"id"`
	ItemType         int    `json:"item_type,omitempty"`
	ItemID           string `json:"item_id,omitempty"`
	Type             int    `json:"type,omitempty,omitempty"`
	CreatedTime      int    `json:"created_time,omitempty"`
	Source           int    `json:"Source,omitempty"`
	BeforeChangeItem string `json:"before_change_item,omitempty"`
}

type tagsResult struct {
	Items   []Tag `json:"items"`
	HasMore bool  `json:"has_more"`
}

type notesResult struct {
	Items   []Note `json:"items"`
	HasMore bool   `json:"has_more"`
}

type notebooksResult struct {
	Items   []Notebook `json:"items"`
	HasMore bool       `json:"has_more"`
}

type Item struct {
	ID       string `json:"id"`
	ParentID string `json:"parent_id"`
	Title    string `json:"title"`
}
type searchResult struct {
	Items   []Item `json:"items"`
	HasMore bool   `json:"has_more"`
}

type CellFormat struct {
	Name   string
	Field  string
	Format string
}

const (
	joplinMinPortNum   = 41184
	joplinMaxPortNum   = 41194
	retriesGetApiToken = 20
)

const (
	ItemTypeName               = "name"
	ItemTypeFolder             = "folder"
	ItemTypeSetting            = "setting"
	ItemTypeResource           = "resource"
	ItemTypeTag                = "tag"
	ItemTypeNoteTag            = "note_tag"
	ItemTypeSearch             = "search"
	ItemTypeAlarm              = "alarm"
	ItemTypeMasterKey          = "master_key"
	ItemTypeItemChange         = "item_change"
	ItemTypeNoteResource       = "note_resource"
	ItemTypeResourceLocalState = "resource_local_state"
	ItemTypeRevision           = "revision"
	ItemTypeMigration          = "migration"
	ItemTypeSmartFilter        = "smart_filter"
	ItemTypeCommand            = "command"
)

type NoteFormat int

const (
	Undefined NoteFormat = iota
	Markdown
	HTML
)

var ItemTypes = []string{
	ItemTypeName,
	ItemTypeFolder,
	ItemTypeSetting,
	ItemTypeResource,
	ItemTypeTag,
	ItemTypeNoteTag,
	ItemTypeSearch,
	ItemTypeAlarm,
	ItemTypeMasterKey,
	ItemTypeItemChange,
	ItemTypeNoteResource,
	ItemTypeResourceLocalState,
	ItemTypeRevision,
	ItemTypeMigration,
	ItemTypeSmartFilter,
	ItemTypeCommand,
}

var TagFormats = map[string]CellFormat{
	"id": {
		"ID",
		"ID",
		"%-32s",
	},
	"parent_id": {
		"Parent ID",
		"ParentID",
		"%-32s",
	},
	"title": {
		"Title",
		"Title",
		"%-60.60s",
	},
	"created_time": {
		"Created Time",
		"CreatedTime",
		"%16.16d",
	},
	"updated_time": {
		"Updated Time",
		"UpdatedTime",
		"%16.16d",
	},
	"user_created_time": {
		"User Created Time",
		"UserCreatedTime",
		"%-16.16d",
	},
	"user_updated_time": {
		"User Updated Time",
		"UserUpdatedTime",
		"%-16.16d",
	},
	"encryption_cipher_text": {
		"Encryption Cipher Text",
		"EncryptionCipherText",
		"%-32.32s",
	},
	"encryption_applied": {
		"Encryption Applied",
		"EncryptionApplied",
		"%-16.16d",
	},
	"is_shared": {
		"Is Shared",
		"IsShared",
		"%-16.16d",
	},
}

var NoteFormats = map[string]CellFormat{
	"id": {
		"ID",
		"ID",
		"%-32s",
	},
	"parent_id": {
		"Parent ID",
		"ParentID",
		"%-32s",
	},
	"title": {
		"Title",
		"Title",
		"%-60.60s",
	},
	"body": {
		"Body",
		"Body",
		"%-60.60s",
	},
	"created_time": {
		"Created Time",
		"CreatedTime",
		"%16.16d",
	},
	"updated_time": {
		"Updated Time",
		"UpdatedTime",
		"%16.16d",
	},
	"is_conflict": {
		"Is Conflict",
		"IsConflict",
		"%-16.16d",
	},
	"latitude": {
		"Latitude",
		"Latitude",
		"%-12.4f",
	},
	"longitude": {
		"Longitude",
		"Longitude",
		"%-12.4f",
	},
	"altitude": {
		"Altitude",
		"Altitude",
		"%-12.4f",
	},
	"author": {
		"Author",
		"Author",
		"%-32.32s",
	},
	"source_url": {
		"Source URL",
		"SourceURL",
		"%-32.32s",
	},
	"is_todo": {
		"Is Todo",
		"IsTodo",
		"%-16.16d",
	},
	"todo_due": {
		"Todo Due",
		"TodoDue",
		"%-16.16d",
	},
	"todo_completed": {
		"Todo Completed",
		"TodoCompleted",
		"%-16.16d",
	},
	"source": {
		"Source",
		"Source",
		"%-32.32s",
	},
	"source_application": {
		"Source Application",
		"SourceApplication",
		"%-32.32s",
	},
	"application_data": {
		"Application Data",
		"ApplicationData",
		"%-32.32s",
	},
	"order": {
		"order",
		"order",
		"%-16.16d",
	},
	"user_created_time": {
		"User Created Time",
		"UserCreatedTime",
		"%-16.16d",
	},
	"user_updated_time": {
		"User Updated Time",
		"UserUpdatedTime",
		"%-16.16d",
	},
	"encryption_cipher_text": {
		"Encryption Cipher Text",
		"EncryptionCipherText",
		"%-32.32s",
	},
	"encryption_applied": {
		"Encryption Applied",
		"EncryptionApplied",
		"%-16.16d",
	},
	"markup_language": {
		"Markup Language",
		"MarkupLanguage",
		"%-16.16d",
	},
	"is_shared": {
		"Is Shared",
		"IsShared",
		"%-16.16d",
	},
	"share_id": {
		"Share ID",
		"ShareID",
		"%-32.32s",
	},
	"conflict_original_id": {
		"Conflict Original ID",
		"ConflictOriginalID",
		"%-32.32s",
	},
	"master_key_id": {
		"Master Key ID",
		"MasterKeyID",
		"%-32.32s",
	},
	"body_html": {
		"Body HTML",
		"BodyHTML",
		"%-32.32s",
	},
	"base_url": {
		"Base URL",
		"BaseURL",
		"%-32.32s",
	},
	"image_data_url": {
		"Image Data URL",
		"ImageDataURL",
		"%-32.32s",
	},
	"crop_rect": {
		"Crop Rect",
		"CropRect",
		"%-32.32s",
	},
}

var ResourceFormats = map[string]CellFormat{
	"id": {
		"ID",
		"ID",
		"%-32s",
	},
	"title": {
		"Title",
		"Title",
		"%-60.60s",
	},
	"mime": {
		"Mime",
		"Mime",
		"%-32.32s",
	},
	"filename": {
		"Filename",
		"Filename",
		"%-32.32s",
	},
	"created_time": {
		"Created Time",
		"CreatedTime",
		"%16.16d",
	},
	"updated_time": {
		"Updated Time",
		"UpdatedTime",
		"%16.16d",
	},
	"user_created_time": {
		"User Created Time",
		"UserCreatedTime",
		"%-16.16d",
	},
	"user_updated_time": {
		"User Updated Time",
		"UserUpdatedTime",
		"%-16.16d",
	},
	"file_extension": {
		"File Extension",
		"FileExtension",
		"%-32.32s",
	},
	"encryption_cipher_text": {
		"Encryption Cipher Text",
		"EncryptionCipherText",
		"%-32.32s",
	},
	"encryption_applied": {
		"Encryption Applied",
		"EncryptionApplied",
		"%-16.16d",
	},
	"encryption_blob_encrypted": {
		"Encryption Blob Encrypted",
		"EncryptionBlobEncrypted",
		"%-16.16d",
	},
	"size": {
		"Size",
		"Size",
		"%-16.16d",
	},
	"is_shared": {
		"Is Shared",
		"IsShared",
		"%-16.16d",
	},
	"share_id": {
		"Share ID",
		"ShareID",
		"%-32.32s",
	},
	"master_key_id": {
		"Master Key ID",
		"MasterKeyID",
		"%-32.32s",
	},
}

var NotebookFormats = map[string]CellFormat{
	"id": {
		"ID",
		"ID",
		"%-32s",
	},
	"parent_id": {
		"Parent ID",
		"ParentID",
		"%-32s",
	},
	"title": {
		"Title",
		"Title",
		"%-60.60s",
	},
	"created_time": {
		"Created Time",
		"CreatedTime",
		"%16.16d",
	},
	"updated_time": {
		"Updated Time",
		"UpdatedTime",
		"%16.16d",
	},
	"user_created_time": {
		"User Created Time",
		"UserCreatedTime",
		"%-16.16d",
	},
	"user_updated_time": {
		"User Updated Time",
		"UserUpdatedTime",
		"%-16.16d",
	},
	"encryption_cipher_text": {
		"Encryption Cipher Text",
		"EncryptionCipherText",
		"%-32.32s",
	},
	"encryption_applied": {
		"Encryption Applied",
		"EncryptionApplied",
		"%-16.16d",
	},
	"is_shared": {
		"Is Shared",
		"IsShared",
		"%-16.16d",
	},
	"share_id": {
		"Share ID",
		"ShareID",
		"%-32.32s",
	},
	"master_key_id": {
		"Master Key ID",
		"MasterKeyID",
		"%-32.32s",
	},
	"icon": {
		"Icon",
		"Icon",
		"%-32.32s",
	},
}

var SearchFormats = map[string]CellFormat{
	"id": {
		"ID",
		"ID",
		"%-32s",
	},
	"parent_id": {
		"Parent ID",
		"ParentID",
		"%-32s",
	},
	"title": {
		"Title",
		"Title",
		"%-60.60s",
	},
}

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
		// Handle response.
		err = fmt.Errorf("got error response, raw dump:\n%s", resp.Dump())

		return token, err
	}

	if resp.IsSuccess() {
		return result.AuthToken, nil
	}

	// Handle response.
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

func (c *Client) GetTag(id string, fields string) (Tag, error) {
	var tag Tag

	resp, err := c.handle.R().
		SetPathParam("id", id).
		SetQueryParam("token", c.apiToken).
		SetQueryParam("fields", fields).
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

	// Handle response.
	err = fmt.Errorf("got unexpected response, raw dump:\n%s", resp.Dump())

	return tag, err
}

func (c *Client) GetNote(id string, fields string) (Note, error) {
	var note Note

	resp, err := c.handle.R().
		SetPathParam("id", id).
		SetQueryParam("token", c.apiToken).
		SetQueryParam("fields", fields).
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
			for _, note := range result.Items {
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

		// Handle response.
		err = fmt.Errorf("got unexpected response, raw dump:\n%s", resp.Dump())

		return notes, err
	}
}

func (c *Client) GetAllNotes(fields string, orderBy string, orderDir string) ([]Note, error) {
	var result notesResult
	var notes []Note

	page := 1

	queryParams := map[string]string{
		"token":  c.apiToken,
		"fields": fields,
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
			for _, note := range result.Items {
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

		// Handle response.
		err = fmt.Errorf("got unexpected response, raw dump:\n%s", resp.Dump())

		return notes, err
	}
}

func (c *Client) GetNotesInNotebook(id string, fields string, orderBy string, orderDir string) ([]Note, error) {
	var result notesResult
	var notes []Note

	page := 1

	queryParams := map[string]string{
		"token":  c.apiToken,
		"fields": fields,
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
			Get(fmt.Sprintf("http://localhost:%d/folders/{id}/notes", c.port))
		if err != nil {
			return notes, err
		}

		if resp.IsError() {
			// handle response.
			err = fmt.Errorf("got error response, raw dump:\n%s", resp.Dump())

			return notes, err
		}

		if resp.IsSuccess() {
			for _, note := range result.Items {
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

		// Handle response.
		err = fmt.Errorf("got unexpected response, raw dump:\n%s", resp.Dump())

		return notes, err
	}
}

func (c *Client) GetAllNotebooks(fields string, orderBy string, orderDir string) ([]Notebook, error) {
	var result notebooksResult
	var notebooks []Notebook

	page := 1

	queryParams := map[string]string{
		"token":  c.apiToken,
		"fields": fields,
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
			Get(fmt.Sprintf("http://localhost:%d/folders", c.port))
		if err != nil {
			return notebooks, err
		}

		if resp.IsError() {
			// Handle response.
			err = fmt.Errorf("got error response, raw dump:\n%s", resp.Dump())

			return notebooks, err
		}

		if resp.IsSuccess() {
			for _, notebook := range result.Items {
				notebooks = append(notebooks, notebook)
			}

			if result.HasMore {
				page++

				queryParams["page"] = strconv.Itoa(page)

				continue
			} else {
				return notebooks, nil
			}
		}

		// Handle response.
		err = fmt.Errorf("got unexpected response, raw dump:\n%s", resp.Dump())

		return notebooks, err
	}
}

func (c *Client) GetNotebook(id string, fields string) (Notebook, error) {
	var notebook Notebook

	resp, err := c.handle.R().
		SetPathParam("id", id).
		SetQueryParam("token", c.apiToken).
		SetQueryParam("fields", fields).
		SetResult(&notebook).
		SetError(&notebook).
		Get(fmt.Sprintf("http://localhost:%d/folders/{id}", c.port))
	if err != nil {
		return notebook, err
	}

	if resp.IsError() {
		if resp.StatusCode == 404 {
			err = fmt.Errorf("could not find notebook with ID '%s'", id)
		} else {
			err = fmt.Errorf("got error response, raw dump:\n%s", resp.Dump())
		}

		return notebook, err
	}

	if resp.IsSuccess() {
		return notebook, nil
	}

	// Handle response.
	err = fmt.Errorf("got unexpected response, raw dump:\n%s", resp.Dump())

	return notebook, err
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
			// Handle response.
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

		// Handle response.
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
		// Handle response.
		err = fmt.Errorf("got error response, raw dump:\n%s", resp.Dump())

		return err
	}

	if resp.IsSuccess() {
		return nil
	}

	// Handle response.
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
		// Handle response.
		err = fmt.Errorf("got error response, raw dump:\n%s", resp.Dump())

		return err
	}

	if resp.IsSuccess() {
		return nil
	}

	// Handle response.
	err = fmt.Errorf("got unexpected response, raw dump:\n%s", resp.Dump())

	return err
}

func (c *Client) Search(query string, queryType string, fields string) ([]Item, error) {
	var result searchResult
	var items []Item

	page := 1

	queryParams := map[string]string{
		"token": c.apiToken,
		"page":  strconv.Itoa(page),
		"query": query,
	}

	if len(queryType) != 0 {
		queryParams["type"] = queryType
	}

	if len(fields) != 0 {
		queryParams["fields"] = fields
	}

	for {
		resp, err := c.handle.R().
			SetQueryParams(queryParams).
			SetResult(&result).
			SetError(&result).
			Get(fmt.Sprintf("http://localhost:%d/search", c.port))
		if err != nil {
			return items, err
		}

		if resp.IsError() {
			// Handle response.
			err = fmt.Errorf("got error response, raw dump:\n%s", resp.Dump())

			return items, err
		}

		if resp.IsSuccess() {
			for _, item := range result.Items {
				items = append(items, item)
			}

			if result.HasMore {
				page++

				queryParams["page"] = strconv.Itoa(page)

				continue
			}

			return items, nil
		}

		// Handle response.
		err = fmt.Errorf("got unexpected response, raw dump:\n%s", resp.Dump())

		return items, err
	}
}

func (c *Client) GetApiToken() string {
	return c.apiToken
}

func (nf NoteFormat) String() string {
	switch nf {
	case Markdown:
		return "Markdown"
	case HTML:
		return "HTML"
	}

	return "unknown"
}

func (c *Client) CreateNote(title string, format NoteFormat, body string, notebook string, tags []string) error {
	if format == Undefined {
		return fmt.Errorf("unknown note format")
	}

	// We've to get the ID of the notebook first.
	items, err := c.Search(notebook, "folder", "")
	if err != nil {
		return err
	}

	if len(items) != 1 {
		return fmt.Errorf("could not find notebook called '%s'", notebook)
	}

	var data map[string]string

	if format == Markdown {
		data = map[string]string{
			"title": title,
			"body":  body,
		}
	} else {
		data = map[string]string{
			"title":     title,
			"body_html": body,
		}
	}

	queryParams := map[string]string{
		"token": c.apiToken,
	}

	resp, err := c.handle.R().
		SetQueryParams(queryParams).
		SetBody(data).
		Post(fmt.Sprintf("http://localhost:%d/notes", c.port))
	if err != nil {
		return err
	}

	if resp.IsError() {
		// Handle response.
		err = fmt.Errorf("got error response:\n%s\n%s", resp.Status, resp.Dump())

		return err
	}

	if resp.IsSuccess() {
		// Ok, we've successfully generated the note, next step is adding the specified tags.

		var note Note

		err := json.Unmarshal(resp.Bytes(), &note)
		if err != nil {
			return err
		}

		for _, tag := range tags {
			items, err := c.Search(tag, "tag", "")
			if err != nil {
				return err
			}

			if len(items) != 1 {
				return fmt.Errorf("could not find tag called '%s'", tag)
			}

			err = c.AddTagToNote(items[0].ID, note)
			if err != nil {
				return err
			}
		}

		return c.MoveNoteToNotebook(note, items[0].ID)
	}

	// Handle response.
	return fmt.Errorf("got unexpected response, raw dump:\n%s", resp.Dump())
}

func (c *Client) MoveNoteToNotebook(note Note, notebook string) error {
	queryParams := map[string]string{
		"token": c.apiToken,
	}

	note.ParentID = notebook

	resp, err := c.handle.R().
		SetPathParam("id", note.ID).
		SetQueryParams(queryParams).
		SetBody(note).
		Put(fmt.Sprintf("http://localhost:%d/notes/{id}", c.port))
	if err != nil {
		return err
	}

	if resp.IsError() {
		// Handle response.
		return fmt.Errorf("got error response:\n%s\n%s", resp.Status, resp.Dump())
	}

	if resp.IsSuccess() {
		return nil
	}

	// Handle response.
	return fmt.Errorf("got unexpected response, raw dump:\n%s", resp.Dump())
}

func (c *Client) AddTagToNote(tagID string, note Note) error {
	queryParams := map[string]string{
		"token": c.apiToken,
	}

	resp, err := c.handle.R().
		SetPathParam("id", tagID).
		SetQueryParams(queryParams).
		SetBody(note).
		Post(fmt.Sprintf("http://localhost:%d/tags/{id}/notes", c.port))
	if err != nil {
		return err
	}

	if resp.IsError() {
		// Handle response.
		return fmt.Errorf("got error response:\n%s\n%s", resp.Status, resp.Dump())
	}

	if resp.IsSuccess() {
		return nil
	}

	// Handle response.
	return fmt.Errorf("got unexpected response, raw dump:\n%s", resp.Dump())
}
