package main

import (
	"context"
	"sort"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

const DnsLabel = "docker_dns.subdomain"

func getExpectedSubdomains(domain string) ([]string, error) {
	expectedDomains := []string{}

	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return expectedDomains, err
	}

	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		return expectedDomains, err
	}

	for _, container := range containers {
		// Is this container enabled as a backend for traefik?
		if value, ok := container.Labels["traefik.enable"]; !ok || value == "false" {
			continue
		}

		// Grab the value of the DNS key
		if value, ok := container.Labels[DnsLabel]; ok {
			expectedDomains = append(expectedDomains, value+"."+domain)
		}

		// Otherwise, this container doesn't get a DNS record
	}

	// Sort it, so comparisons work easier
	sort.Strings(expectedDomains)
	return expectedDomains, nil
}
