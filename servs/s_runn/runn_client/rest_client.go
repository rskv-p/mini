package runn_client

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"
)

// RESTClient handles HTTP requests to the server
type RESTClient struct {
	BaseURL string       // Base URL for the API (e.g., http://localhost:8080/api)
	Client  *http.Client // HTTP client for making requests
}

// NewRESTClient creates a new REST client with the provided base URL
func NewRESTClient(baseURL string) *RESTClient {
	return &RESTClient{
		BaseURL: baseURL,
		Client:  &http.Client{},
	}
}

// List retrieves a list of processes from the server
func (c *RESTClient) List() ([]*Process, error) {
	resp, err := c.Client.Get(c.BaseURL + "/processes")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var procs []*Process
	if err := json.NewDecoder(resp.Body).Decode(&procs); err != nil {
		return nil, err
	}
	return procs, nil
}

// Start sends a request to start the process with the given ID
func (c *RESTClient) Start(id uint64) error {
	_, err := c.Client.Post(c.BaseURL+"/start/"+strconv.FormatUint(id, 10), "", nil)
	return err
}

// Stop sends a request to stop the process with the given ID
func (c *RESTClient) Stop(id uint64) error {
	_, err := c.Client.Post(c.BaseURL+"/stop/"+strconv.FormatUint(id, 10), "", nil)
	return err
}

// Pause sends a request to pause the process with the given ID
func (c *RESTClient) Pause(id uint64) error {
	_, err := c.Client.Post(c.BaseURL+"/pause/"+strconv.FormatUint(id, 10), "", nil)
	return err
}

// Resume sends a request to resume the process with the given ID
func (c *RESTClient) Resume(id uint64) error {
	_, err := c.Client.Post(c.BaseURL+"/resume/"+strconv.FormatUint(id, 10), "", nil)
	return err
}

// Remove sends a request to remove the process with the given ID
func (c *RESTClient) Remove(id uint64) error {
	req, err := http.NewRequest(http.MethodDelete, c.BaseURL+"/remove/"+strconv.FormatUint(id, 10), nil)
	if err != nil {
		return err
	}
	_, err = c.Client.Do(req)
	return err
}

// Add sends a request to add a new process with the given command and directory
func (c *RESTClient) Add(cmd, dir string) (*Process, error) {
	body := map[string]string{"cmd": cmd, "dir": dir}
	buf, _ := json.Marshal(body)

	resp, err := c.Client.Post(c.BaseURL+"/add", "application/json", bytes.NewReader(buf))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var proc Process
	if err := json.NewDecoder(resp.Body).Decode(&proc); err != nil {
		return nil, err
	}
	return &proc, nil
}
