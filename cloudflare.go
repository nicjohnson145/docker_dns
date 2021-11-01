package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/tidwall/gjson"
)

const cloudflare = "cloudflare"

type Cloudflare struct {
	username string
	password string
	domain   string
	apiRoot  string
	zoneId   string
}

type record struct {
	Type    string
	ID      string
	Name    string
	Content string
}

type cloudflareRecord struct {
	Type    string `json:"type"`
	Name    string `json:"name"`
	Content string `json:"content"`
	TTL     int    `json:"ttl"`
	Proxied bool   `json:"proxied"`
}

var _ Provider = (*Cloudflare)(nil)

func (c *Cloudflare) Load(settings Settings) {
	c.username = settings.Username
	c.password = settings.Password
	c.domain = settings.Domain
	c.apiRoot = "https://api.cloudflare.com/client/v4"
}

func (c *Cloudflare) Reconcile(entries []string) error {
	aRecords, txtRecords, err := c.getExistingRecords()
	if err != nil {
		return err
	}

	// Add any new records
	for _, domain := range entries {
		if _, ok := aRecords[domain]; !ok {
			err := c.createRecord(domain)
			if err != nil {
				return err
			}
		}
	}

	// Remove any docker-dns managed records no longer needed
	for _, rec := range aRecords {
		noContainer := !stringInSlice(rec.Name, entries)
		_, hasTxt := txtRecords[rec.Name]

		if noContainer && hasTxt {
			aRecord := aRecords[rec.Name]
			txtRecord := txtRecords[rec.Name]
			err := c.deleteRecord(aRecord)
			if err != nil {
				return err
			}
			err = c.deleteRecord(txtRecord)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *Cloudflare) getExistingRecords() (map[string]record, map[string]record, error) {
	aRecords := make(map[string]record)
	txtRecords := make(map[string]record)

	id, err := c.getZoneId()
	if err != nil {
		return aRecords, txtRecords, err
	}

	url := strings.Join([]string{c.apiRoot, "zones", id, "dns_records"}, "/")
	resp, err := getJsonResponse("GET", url, c.authHeaders(), nil)

	if err != nil {
		return aRecords, txtRecords, err
	}

	records := gjson.Get(resp, "result")
	records.ForEach(func(key, value gjson.Result) bool {
		record := record{
			Type:    value.Get("type").String(),
			ID:      value.Get("id").String(),
			Name:    value.Get("name").String(),
			Content: value.Get("content").String(),
		}
		if record.Type == "A" {
			aRecords[record.Name] = record
		} else if record.Type == "TXT" {
			txtRecords[record.Name] = record
		}

		return true
	})

	return aRecords, txtRecords, nil
}

func (c *Cloudflare) authHeaders() map[string]string {
	return map[string]string{
		"X-Auth-Email": c.username,
		"X-Auth-Key":   c.password,
	}
}

func (c *Cloudflare) getZoneId() (string, error) {
	if c.zoneId != "" {
		return c.zoneId, nil
	}

	query := url.Values{
		"name":     {c.domain},
		"per_page": {"50"},
	}
	url := c.apiRoot + "/zones?" + query.Encode()
	resp, err := getJsonResponse("GET", url, c.authHeaders(), nil)
	if err != nil {
		return "", err
	}

	id := gjson.Get(resp, "result.0.id")
	if !id.Exists() {
		return "", fmt.Errorf("could not retrieve zone id")
	}

	c.zoneId = id.String()
	return c.zoneId, nil
}

func (c *Cloudflare) createRecord(name string) error {
	ip, err := getCurrentIp()
	if err != nil {
		return fmt.Errorf("getting current ip: %w", err)
	}

	zoneId, err := c.getZoneId()
	if err != nil {
		return fmt.Errorf("getting zone id: %w", err)
	}

	aBody, err := json.Marshal(cloudflareRecord{
		Type:    "A",
		Name:    name,
		Content: ip,
		TTL:     1, // cloudflare auto ttl
		Proxied: true,
	})
	if err != nil {
		return fmt.Errorf("creating a record body: %w", err)
	}
	txtBody, err := json.Marshal(cloudflareRecord{
		Type:    "TXT",
		Name:    name,
		Content: "owner=docker_dns",
		TTL:     1, // cloudflare auto ttl
		Proxied: false,
	})
	if err != nil {
		return fmt.Errorf("creating txt record body: %w", err)
	}

	url := c.apiRoot + fmt.Sprintf("/zones/%v/dns_records", zoneId)
	_, err = getJsonResponse("POST", url, c.authHeaders(), bytes.NewBuffer(aBody))
	if err != nil {
		return fmt.Errorf("creating a record: %w", err)
	}

	_, err = getJsonResponse("POST", url, c.authHeaders(), bytes.NewBuffer(txtBody))
	if err != nil {
		return fmt.Errorf("creating txt record: %w", err)
	}

	return nil
}

func (c *Cloudflare) deleteRecord(rec record) error {
	zoneId, err := c.getZoneId()
	if err != nil {
		return fmt.Errorf("getting zone id: %w", err)
	}

	url := c.apiRoot + fmt.Sprintf("/zones/%v/dns_records/%v", zoneId, rec.ID)
	_, err = getJsonResponse("DELETE", url, c.authHeaders(), nil)
	return err
}
