package whois

import (
	"bufio"
	"fmt"
	"net"
	"regexp"
	"strings"
	"time"
)

const (
	ianaServer  = "whois.iana.org:43"
	queryTimout = 10 * time.Second
)

type Client struct {
	Debug    func(string)
	Override string
}

func NewClient(debug func(string)) *Client {
	return &Client{Debug: debug}
}

func (c *Client) Query(domain string) (string, error) {
	server := c.Override
	if server == "" || server == "0" {
		var err error
		server, err = c.resolveServer(domain)
		if err != nil {
			return "", err
		}
	}

	if c.Debug != nil {
		c.Debug(fmt.Sprintf("Running WHOIS for %s with server %s", domain, server))
	}

	output, err := c.queryServer(server, domain)
	if err != nil {
		return "", err
	}

	if strings.TrimSpace(output) == "" {
		return "", fmt.Errorf("empty WHOIS response")
	}

	lower := strings.ToLower(output)
	if strings.Contains(lower, "no match for") ||
		strings.Contains(lower, "not found") ||
		strings.Contains(lower, "no domain") {
		return "", fmt.Errorf("domain %s doesn't exist", domain)
	}

	if strings.Contains(lower, "query rate limit exceeded") ||
		strings.Contains(lower, "whois_limit_exceeded") {
		return "", fmt.Errorf("rate limited WHOIS")
	}

	if strings.Contains(lower, "connection refused") ||
		strings.Contains(lower, "timeout") ||
		strings.Contains(lower, "no whois server") ||
		strings.Contains(lower, "socket") ||
		strings.Contains(lower, "fgets") {
		return "", fmt.Errorf("WHOIS query failed for %s", domain)
	}

	return output, nil
}

func (c *Client) resolveServer(domain string) (string, error) {
	tld := domain[strings.LastIndex(domain, ".")+1:]

	response, err := c.queryServer("whois.iana.org", tld)
	if err != nil {
		return "", err
	}

	server := parseReferral(response)
	if server != "" {
		return server, nil
	}

	response, err = c.queryServer("whois.iana.org", domain)
	if err != nil {
		return "", err
	}

	return parseReferral(response), nil
}

var referralPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?im)^refer:\s*(\S+)`),
	regexp.MustCompile(`(?im)^whois:\s*(\S+)`),
}

func parseReferral(response string) string {
	for _, pattern := range referralPatterns {
		if match := pattern.FindStringSubmatch(response); len(match) > 1 {
			return strings.TrimSpace(match[1])
		}
	}
	return ""
}

func (c *Client) queryServer(server, query string) (string, error) {
	addr := server
	if !strings.Contains(addr, ":") {
		addr = net.JoinHostPort(addr, "43")
	}

	conn, err := net.DialTimeout("tcp", addr, queryTimout)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	_ = conn.SetDeadline(time.Now().Add(queryTimout))

	if _, err := fmt.Fprintf(conn, "%s\r\n", query); err != nil {
		return "", err
	}

	var builder strings.Builder
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		builder.WriteString(scanner.Text())
		builder.WriteByte('\n')
	}

	if err := scanner.Err(); err != nil {
		return builder.String(), err
	}

	return builder.String(), nil
}