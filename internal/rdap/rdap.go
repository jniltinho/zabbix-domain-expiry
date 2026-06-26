package rdap

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const bootstrapURL = "https://data.iana.org/rdap/dns.json"

type Client struct {
	HTTP     *http.Client
	Debug    func(string)
	Override string
}

type bootstrapFile struct {
	Services [][]json.RawMessage `json:"services"`
}

type rdapResponse struct {
	Events []struct {
		EventAction string `json:"eventAction"`
		EventDate   string `json:"eventDate"`
	} `json:"events"`
}

func NewClient(debug func(string)) *Client {
	return &Client{
		HTTP: &http.Client{
			Timeout: 5 * time.Second,
		},
		Debug: debug,
	}
}

func (c *Client) GetExpiration(domain string) (string, error) {
	tld := domain[strings.LastIndex(domain, ".")+1:]
	server := c.Override

	if server == "" || server == "0" {
		var err error
		server, err = c.lookupServer(tld)
		if err != nil {
			return "", err
		}
		if server == "" {
			return "", fmt.Errorf("no RDAP server for TLD .%s", tld)
		}
	}

	server = adjustURL(tld, server)
	if c.Debug != nil {
		c.Debug(fmt.Sprintf("Using RDAP server: %s", server))
	}

	url := strings.TrimRight(server, "/") + "/domain/" + domain
	if c.Debug != nil {
		c.Debug(fmt.Sprintf("Querying RDAP URL: %s", url))
	}

	body, err := c.fetchWithRetry(url, 3)
	if err != nil {
		return "", err
	}

	var resp rdapResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return "", err
	}

	for _, event := range resp.Events {
		if strings.Contains(strings.ToLower(event.EventAction), "expiration") && event.EventDate != "" {
			return strings.Split(event.EventDate, "T")[0], nil
		}
	}

	return "", fmt.Errorf("no expiration date in RDAP response")
}

func adjustURL(tld, server string) string {
	switch tld {
	case "uk":
		if !strings.HasSuffix(server, "/uk/") {
			return strings.TrimRight(server, "/") + "/uk/"
		}
	}
	return server
}

func (c *Client) lookupServer(tld string) (string, error) {
	if c.Debug != nil {
		c.Debug(fmt.Sprintf("Attempting IANA RDAP lookup for TLD .%s", tld))
	}

	body, err := c.fetchWithRetry(bootstrapURL, 3)
	if err != nil {
		return "", err
	}

	var bootstrap bootstrapFile
	if err := json.Unmarshal(body, &bootstrap); err != nil {
		return "", err
	}

	for _, service := range bootstrap.Services {
		if len(service) < 2 {
			continue
		}

		var tlds []string
		if err := json.Unmarshal(service[0], &tlds); err != nil {
			continue
		}

		for _, entry := range tlds {
			if entry == tld {
				var servers []string
				if err := json.Unmarshal(service[1], &servers); err != nil || len(servers) == 0 {
					continue
				}
				if c.Debug != nil {
					c.Debug(fmt.Sprintf("RDAP server for TLD .%s: %s", tld, servers[0]))
				}
				return servers[0], nil
			}
		}
	}

	return "", nil
}

func (c *Client) fetchWithRetry(url string, retries int) ([]byte, error) {
	var lastErr error
	for attempt := 0; attempt < retries; attempt++ {
		if attempt > 0 {
			time.Sleep(200 * time.Millisecond)
		}

		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Accept", "application/rdap+json, application/json")

		resp, err := c.HTTP.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		body, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()

		if readErr != nil {
			lastErr = readErr
			continue
		}

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return body, nil
		}

		lastErr = fmt.Errorf("HTTP %d from %s", resp.StatusCode, url)
	}

	return nil, lastErr
}