package porkbun

import (
	"net/http"
	"sync"
)

type Client struct {
	apiKey          string
	secretKey       string
	BaseURL         string
	HTTPClient      *http.Client
	mu              sync.Mutex
	recordsCache    map[string][]DnsRecord
	pricingCache    map[string]TldPricing
	glueRecordCache map[string]map[string][]string
	dnssecCache     map[string][]DnssecRecord
	domainListCache []DomainListing
}

type Auth struct {
	APIKey    string `json:"apikey"`
	SecretKey string `json:"secretapikey"`
}

type PingResponse struct {
	Status string `json:"status"`
	YourIP string `json:"yourIp"`
}

type TldPricing struct {
	SLA             float64 `json:"sla"`
	Registration    string  `json:"registration"`
	Renewal         string  `json:"renewal"`
	Transfer        string  `json:"transfer"`
	GraceDayRenewal string  `json:"renew_grace_day"`
}

type PricingResponse struct {
	Status  string                `json:"status"`
	Pricing map[string]TldPricing `json:"pricing"`
}

type NsResponse struct {
	Status string   `json:"status"`
	NS     []string `json:"ns"`
}

type GlueRecordIPs struct {
	V4 []string `json:"v4"`
	V6 []string `json:"v6"`
}

type GetGlueRecordsResponse struct {
	Status string          `json:"status"`
	Hosts  [][]interface{} `json:"hosts"`
}

type DnssecRecord struct {
	Algorithm  string `json:"algorithm"`
	DigestType string `json:"digestType"`
	KeyTag     string `json:"keyTag"`
	Digest     string `json:"digest"`
}

type GetDnssecResponse struct {
	Status    string         `json:"status"`
	DsRecords []DnssecRecord `json:"dsRecords"`
}

type DnsRecord struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	Content string `json:"content"`
	TTL     string `json:"ttl"`
	Prio    string `json:"prio,omitempty"`
	Notes   string `json:"notes,omitempty"`
}

type DomainListing struct {
	Domain       string      `json:"domain"`
	Status       string      `json:"status"`
	Tld          string      `json:"tld"`
	CreateDate   string      `json:"createDate"`
	ExpireDate   string      `json:"expireDate"`
	SecurityLock interface{} `json:"securityLock"`
	WhoisPrivacy interface{} `json:"whoisPrivacy"`
	AutoRenew    interface{} `json:"autoRenew"`
}

type ListAllResponse struct {
	Status  string          `json:"status"`
	Domains []DomainListing `json:"domains"`
}
