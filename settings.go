package main

import (
	"fmt"
	"os"
	"strconv"

	"gopkg.in/ini.v1"
)

type Settings struct {
	Domain       string
	ProviderName string
	Username     string
	Password     string
	Interval     int
}

func LoadSettings() Settings {
	// Domain is required
	var domain string
	if val, ok := os.LookupEnv("DOMAIN"); !ok {
		ExitErr("DOMAIN is required")
	} else {
		domain = val
	}

	// Provider is required
	var provider string
	if val, ok := os.LookupEnv("PROVIDER"); !ok {
		ExitErr("PROVIDER is required")
	} else {
		switch val {
		case cloudflare:
			provider = val
		default:
			ExitErr(fmt.Sprintf("Unknown dns provider %s", val))
		}
	}

	// Load auth details
	var authLocation string
	if val, ok := os.LookupEnv("AUTH_FILE"); !ok {
		authLocation = "/auth/credentials.ini"
	} else {
		authLocation = val
	}
	creds, err := ini.Load(authLocation)
	if err != nil {
		ExitErr(fmt.Errorf("Error loading credentials: %w", err).Error())
	}
	username := creds.Section("").Key("username").String()
	if username == "" {
		ExitErr(fmt.Sprintf("'username' key not found in auth file %s", authLocation))
	}
	password := creds.Section("").Key("password").String()
	if password == "" {
		ExitErr(fmt.Sprintf("'password' key not found in auth file %s", authLocation))
	}

	// Load loop interval
	var loopInterval int
	if val, ok := os.LookupEnv("LOOP_INTERVAL"); !ok {
		loopInterval = 5
	} else {
		intv, err := strconv.Atoi(val)
		if err != nil {
			ExitErr(fmt.Sprintf("Error parsing LOOP_INTERVAL: %v", err))
		}
		loopInterval = intv
	}

	return Settings{
		Domain:       domain,
		ProviderName: provider,
		Username:     username,
		Password:     password,
		Interval:     loopInterval,
	}
}
