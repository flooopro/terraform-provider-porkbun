package porkbun

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const defaultBaseURL = "https://api.porkbun.com/api/json/v3"

func NewClient(apiKey, secretKey string) *Client {
	return &Client{
		apiKey:          apiKey,
		secretKey:       secretKey,
		BaseURL:         defaultBaseURL,
		HTTPClient:      &http.Client{},
		recordsCache:    make(map[string][]DnsRecord),
		pricingCache:    make(map[string]TldPricing),
		glueRecordCache: make(map[string]map[string][]string),
		dnssecCache:     make(map[string][]DnssecRecord),
		domainListCache: nil,
	}
}

func (c *Client) clearDomainCache(domain string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.recordsCache, domain)
	delete(c.glueRecordCache, domain)
	delete(c.dnssecCache, domain)
}

func (c *Client) newAuthenticatedRequest(method, url string, body interface{}) (*http.Request, error) {
	authBody := make(map[string]interface{})
	if body != nil {
		b, _ := json.Marshal(body)
		json.Unmarshal(b, &authBody)
	}
	authBody["apikey"] = c.apiKey
	authBody["secretapikey"] = c.secretKey

	rb, err := json.Marshal(authBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(rb))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Terraform-Provider-Porkbun/0.0.1-SNAPSHOT")
	return req, nil
}

func (c *Client) do(req *http.Request, v interface{}) error {
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error: status code %d, response: %s", resp.StatusCode, string(bodyBytes))
	}

	var statusResponse struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if err := json.Unmarshal(bodyBytes, &statusResponse); err != nil {
		return fmt.Errorf("failed to decode status response: %w", err)
	}

	if statusResponse.Status == "ERROR" {
		return fmt.Errorf("Porkbun API error: %s", statusResponse.Message)
	}

	if v != nil {
		if err := json.Unmarshal(bodyBytes, v); err != nil {
			return fmt.Errorf("failed to decode response body into target: %w", err)
		}
	}

	return nil
}

func (c *Client) CreateRecord(domain string, record DnsRecord) (string, error) {
	url := fmt.Sprintf("%s/dns/create/%s", c.BaseURL, domain)
	req, err := c.newAuthenticatedRequest("POST", url, record)
	if err != nil {
		return "", err
	}

	var response struct {
		Status string `json:"status"`
		ID     int    `json:"id"`
	}

	if err := c.do(req, &response); err != nil {
		return "", err
	}
	c.clearDomainCache(domain)
	return fmt.Sprintf("%d", response.ID), nil
}

func (c *Client) RetrieveRecords(domain string) ([]DnsRecord, error) {
	c.mu.Lock()
	if cachedRecords, found := c.recordsCache[domain]; found {
		c.mu.Unlock()
		return cachedRecords, nil
	}
	c.mu.Unlock()

	url := fmt.Sprintf("%s/dns/retrieve/%s", c.BaseURL, domain)
	req, err := c.newAuthenticatedRequest("POST", url, nil)
	if err != nil {
		return nil, err
	}

	var response struct {
		Status  string      `json:"status"`
		Records []DnsRecord `json:"records"`
	}

	if err := c.do(req, &response); err != nil {
		return nil, err
	}

	c.mu.Lock()
	c.recordsCache[domain] = response.Records
	c.mu.Unlock()

	return response.Records, nil
}

func (c *Client) DeleteRecord(domain, recordID string) error {
	url := fmt.Sprintf("%s/dns/delete/%s/%s", c.BaseURL, domain, recordID)
	req, err := c.newAuthenticatedRequest("POST", url, nil)
	if err != nil {
		return err
	}
	c.clearDomainCache(domain)
	return c.do(req, nil)
}

func (c *Client) EditRecord(domain, recordID string, record DnsRecord) error {
	url := fmt.Sprintf("%s/dns/edit/%s/%s", c.BaseURL, domain, recordID)

	payload := map[string]string{
		"name":    record.Name,
		"type":    record.Type,
		"content": record.Content,
		"ttl":     record.TTL,
	}

	recordType := strings.ToUpper(record.Type)
	if (recordType == "MX" || recordType == "SRV") && record.Prio != "" {
		payload["prio"] = record.Prio
	}

	req, err := c.newAuthenticatedRequest("POST", url, payload)
	if err != nil {
		return err
	}
	c.clearDomainCache(domain)
	return c.do(req, nil)

}

func (c *Client) Ping(ctx context.Context) (string, error) {
	url := fmt.Sprintf("%s/ping", c.BaseURL)
	req, err := c.newAuthenticatedRequest("POST", url, nil)
	if err != nil {
		return "", err
	}

	var response PingResponse
	if err := c.do(req, &response); err != nil {
		return "", err
	}

	return response.YourIP, nil
}

func (c *Client) GetPricing() (map[string]TldPricing, error) {
	c.mu.Lock()
	if len(c.pricingCache) > 0 {
		c.mu.Unlock()
		return c.pricingCache, nil
	}
	c.mu.Unlock()

	url := fmt.Sprintf("%s/pricing/get", c.BaseURL)
	req, err := c.newAuthenticatedRequest("POST", url, nil)
	if err != nil {
		return nil, err
	}

	var response PricingResponse
	if err := c.do(req, &response); err != nil {
		return nil, err
	}

	c.mu.Lock()
	c.pricingCache = response.Pricing
	c.mu.Unlock()

	return response.Pricing, nil
}

func (c *Client) UpdateNameservers(domain string, nameservers []string) error {
	url := fmt.Sprintf("%s/domain/updateNs/%s", c.BaseURL, domain)

	payload := map[string][]string{
		"ns": nameservers,
	}

	req, err := c.newAuthenticatedRequest("POST", url, payload)
	if err != nil {
		return err
	}

	c.clearDomainCache(domain)
	return c.do(req, nil)
}

func (c *Client) GetNameservers(domain string) ([]string, error) {
	url := fmt.Sprintf("%s/domain/getNs/%s", c.BaseURL, domain)

	req, err := c.newAuthenticatedRequest("POST", url, nil)
	if err != nil {
		return nil, err
	}

	var response NsResponse
	if err := c.do(req, &response); err != nil {
		return nil, err
	}

	return response.NS, nil
}

func (c *Client) AddGlueRecord(domain, host string, ips []string) error {
	url := fmt.Sprintf("%s/domain/createGlue/%s/%s", c.BaseURL, domain, host)
	payload := map[string]interface{}{
		"ips": ips,
	}
	req, err := c.newAuthenticatedRequest("POST", url, payload)
	if err != nil {
		return err
	}

	err = c.do(req, nil)
	if err == nil {
		c.clearDomainCache(domain)
	}
	return err
}

func (c *Client) DeleteGlueRecord(domain, host string) error {
	url := fmt.Sprintf("%s/domain/deleteGlue/%s/%s", c.BaseURL, domain, host)

	req, err := c.newAuthenticatedRequest("POST", url, nil)
	if err != nil {
		return err
	}

	err = c.do(req, nil)
	if err == nil {
		c.clearDomainCache(domain)
	}
	return err
}

func (c *Client) GetGlueRecords(domain string) (map[string][]string, error) {
	c.mu.Lock()
	if cachedRecords, found := c.glueRecordCache[domain]; found {
		c.mu.Unlock()
		return cachedRecords, nil
	}
	c.mu.Unlock()

	url := fmt.Sprintf("%s/domain/getGlue/%s", c.BaseURL, domain)
	req, err := c.newAuthenticatedRequest("POST", url, nil)
	if err != nil {
		return nil, err
	}

	var response GetGlueRecordsResponse
	if err := c.do(req, &response); err != nil {
		return nil, err
	}

	parsedRecords := make(map[string][]string)
	for _, hostData := range response.Hosts {
		if len(hostData) != 2 {
			continue
		}
		fqdn, ok := hostData[0].(string)
		if !ok {
			continue
		}
		host := strings.TrimSuffix(fqdn, "."+domain)

		ipMapBytes, err := json.Marshal(hostData[1])
		if err != nil {
			continue
		}
		var ips GlueRecordIPs
		if err := json.Unmarshal(ipMapBytes, &ips); err != nil {
			continue
		}

		allIPs := append(ips.V4, ips.V6...)
		parsedRecords[host] = allIPs
	}

	c.mu.Lock()
	c.glueRecordCache[domain] = parsedRecords
	c.mu.Unlock()

	return parsedRecords, nil
}

func (c *Client) GetDnssecRecords(domain string) ([]DnssecRecord, error) {
	c.mu.Lock()
	if cachedRecords, found := c.dnssecCache[domain]; found {
		c.mu.Unlock()
		return cachedRecords, nil
	}
	c.mu.Unlock()

	url := fmt.Sprintf("%s/dns/getDnssec/%s", c.BaseURL, domain)
	req, err := c.newAuthenticatedRequest("POST", url, nil)
	if err != nil {
		return nil, err
	}

	var response GetDnssecResponse
	if err := c.do(req, &response); err != nil {
		return nil, err
	}

	c.mu.Lock()
	c.dnssecCache[domain] = response.DsRecords
	c.mu.Unlock()

	return response.DsRecords, nil
}

func (c *Client) AddDnssecRecord(domain string, record DnssecRecord) error {
	url := fmt.Sprintf("%s/dns/addDnssec/%s", c.BaseURL, domain)
	req, err := c.newAuthenticatedRequest("POST", url, record)
	if err != nil {
		return err
	}

	err = c.do(req, nil)
	if err == nil {
		c.clearDomainCache(domain)
	}
	return err
}

func (c *Client) DeleteDnssecRecord(domain string, record DnssecRecord) error {
	url := fmt.Sprintf("%s/dns/deleteDnssec/%s", c.BaseURL, domain)
	req, err := c.newAuthenticatedRequest("POST", url, record)
	if err != nil {
		return err
	}

	err = c.do(req, nil)
	if err == nil {
		c.clearDomainCache(domain)
	}
	return err
}

func (c *Client) ListAllDomains() ([]DomainListing, error) {
	c.mu.Lock()
	if c.domainListCache != nil {
		c.mu.Unlock()
		return c.domainListCache, nil
	}
	c.mu.Unlock()

	url := fmt.Sprintf("%s/domain/listAll", c.BaseURL)
	req, err := c.newAuthenticatedRequest("POST", url, nil)
	if err != nil {
		return nil, err
	}

	var response ListAllResponse
	if err := c.do(req, &response); err != nil {
		return nil, err
	}

	c.mu.Lock()
	c.domainListCache = response.Domains
	c.mu.Unlock()

	return response.Domains, nil
}
