package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	registryName   string
	imageLimit     int
	clientId       string
	clientSecret   string
	verboseLogging bool
	dryRun         bool
}

func (c *Config) Setup() {

	// flags for registry name
	flag.StringVar(&c.registryName, "registry", "", "Name of your Azure Container Registry")
	flag.StringVar(&c.registryName, "r", "", "Name of your Azure Container Registry (shorthand)")

	// flags for image limit
	flag.IntVar(&c.imageLimit, "limit", 0, "Number of tags you'd like to keep for each image")
	flag.IntVar(&c.imageLimit, "l", 0, "Number of tags you'd like to keep for each image (shorthand)")

	// flags for client ID
	flag.StringVar(&c.clientId, "user", "", "Client ID / User for authentication")
	flag.StringVar(&c.clientId, "u", "", "Client ID / User for authentication (shorthand)")

	// flags for client secret
	flag.StringVar(&c.clientSecret, "password", "", "Secret / Password for authentication")
	flag.StringVar(&c.clientSecret, "p", "", "Client ID/User for authentication (shorthand)")

	// flags for logging
	flag.BoolVar(&c.verboseLogging, "v", false, "Enable verbose logging")

	// flags for dry run
	flag.BoolVar(&c.dryRun, "d", false, "Perform a dry run")
}

func (c *Config) IsValid() bool {

	if c.registryName == "" {
		if os.Getenv("DUSTY_REG_NAME") != "" {
			c.registryName = os.Getenv("DUSTY_REG_NAME")
		} else {
			return false
		}
	}

	if c.clientId == "" {
		if os.Getenv("DUSTY_CLIENT_ID") != "" {
			c.clientId = os.Getenv("DUSTY_CLIENT_ID")
		} else {
			return false
		}
	}

	if c.clientSecret == "" {
		if os.Getenv("DUSTY_CLIENT_SECRET") != "" {
			c.clientSecret = os.Getenv("DUSTY_CLIENT_SECRET")
		} else {
			return false
		}
	}

	if c.imageLimit <= 0 {
		if os.Getenv("DUSTY_TAG_LIMIT") != "" {
			i, err := strconv.Atoi(os.Getenv("DUSTY_TAG_LIMIT"))
			if err != nil {
				log.Fatal("Error converting DUSTY_TAG_LIMIT to integer value")
			} else {
				c.imageLimit = i
			}
		} else {
			return false
		}
	}

	// If they've not given a full URL, assume the standard azurecr URL
	if !strings.HasPrefix(c.registryName, "http") {
		c.registryName = fmt.Sprintf("https://%s.azurecr.io:443", c.registryName)
	}

	return true
}
