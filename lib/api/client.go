package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/arv28/form3-accountapi-client/lib/accounts"
)

var (
	// default timeout for client request
	timeout = 10 * time.Second

	// API Path for accounts resource
	ACCOUNTS_API_PATH = "/v1/organisation/accounts"

	// backoff schedule on how often to retry failed http requests
	backoffSchedule = []time.Duration{
		2 * time.Second,
		3 * time.Second,
		5 * time.Second,
	}
)

// Define Client
type Client struct {
	HostURL    string
	HTTPClient *http.Client
}

func NewClient(hostURL string) *Client {
	return &Client{
		HostURL: hostURL,
		HTTPClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// Fetch: Get a single account using the account ID
func (c *Client) Fetch(id string) (*accounts.AccountData, error) {
	if id == "" {
		return nil, errors.New("Invalid account id")
	}

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s%s/%s", c.HostURL, ACCOUNTS_API_PATH, id), nil)
	if err != nil {
		return nil, err
	}

	res := accounts.AccountDataResponse{}
	if err := c.sendRequest(req, &res); err != nil {
		return nil, err
	}

	return res.Data, nil
}

// Create: Register an existing bank account with Form3 or create a new one
func (c *Client) Create(accountData *accounts.AccountData) (*accounts.AccountData, error) {
	req := accounts.AccountDataRequest{
		Data: accountData,
	}
	// encode the request data
	requestBody, err := json.Marshal(req)
	if err != nil {
		log.Fatalln(err)
		return nil, err
	}

	// do POST
	createReq, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s%s", c.HostURL, ACCOUNTS_API_PATH), bytes.NewBuffer(requestBody))
	if err != nil {
		log.Fatalln(err)
		return nil, err
	}

	createResp := accounts.AccountDataResponse{}
	if err := c.sendRequest(createReq, &createResp); err != nil {
		return nil, err
	}

	return createResp.Data, nil
}

// Delete: Delete a form3 account
func (c *Client) Delete(accountId string, version int) error {
	deleteReq, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s%s/%s?version=%d", c.HostURL, ACCOUNTS_API_PATH, accountId, version), nil)
	if err != nil {
		log.Fatalln(err)
		return err
	}

	if err := c.sendRequest(deleteReq, nil); err != nil {
		return err
	}

	return nil
}

// sendRequest: sendRequest is the general method called by exported methods in client library
// to actually perform http client requests.
func (c *Client) sendRequest(req *http.Request, v interface{}) error {
	req.Header.Set("Content-Type", "application/vnd.api+json")
	req.Header.Set("Accept", "application/vnd.api+json")

	var err error
	var res *http.Response

	// Retry with backoff schedule
	for _, backoff := range backoffSchedule {
		res, err = c.HTTPClient.Do(req)

		if err == nil {
			break
		}

		fmt.Printf("Request error: %+v\n", err)
		fmt.Printf("Retrying in %v\n", backoff)
		if res != nil {
			_ = res.Body.Close()
		}
		time.Sleep(backoff)
	}

	// All retries failed
	if err != nil {
		return err
	}

	// Resource leak if response body isn't closed
	defer res.Body.Close()

	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusBadRequest {
		data := getErrorType(res.StatusCode)
		if err := json.NewDecoder(res.Body).Decode(&data); err == nil {
			return data
		}
		//fmt.Println("err: ", err)
		return err
	}

	if v != nil {
		if err = json.NewDecoder(res.Body).Decode(&v); err != nil {
			return err
		}
	}

	return nil
}

// Error type factory
func getErrorType(code int) error {
	switch {
	case code == http.StatusBadRequest:
		return &BadRequestError{}
	case code == http.StatusNotFound:
		return &NotFoundError{}
	case code == http.StatusConflict:
		return &ConflictError{}
	case code >= http.StatusInternalServerError:
		return &InternalServerError{}
	default:
		return errors.New("Unknown code")

	}
}
