package common

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/hashicorp/packer/template/interpolate"
)

// ISOConfig contains configuration for downloading ISO images.
type ISOConfig struct {
	ISOChecksum     string   `mapstructure:"iso_checksum"`
	ISOChecksumURL  string   `mapstructure:"iso_checksum_url"`
	ISOChecksumType string   `mapstructure:"iso_checksum_type"`
	ISOUrls         []string `mapstructure:"iso_urls"`
	TargetPath      string   `mapstructure:"iso_target_path"`
	TargetExtension string   `mapstructure:"iso_target_extension"`
	RawSingleISOUrl string   `mapstructure:"iso_url"`
}

func (c *ISOConfig) Prepare(ctx *interpolate.Context) (warnings []string, errs []error) {

	if c.RawSingleISOUrl != "" {
		// make sure only array is set
		c.ISOUrls = append([]string{c.RawSingleISOUrl}, c.ISOUrls...)
		c.RawSingleISOUrl = ""
	}
	if len(c.ISOUrls) == 0 {
		errs = append(
			errs, errors.New("One of iso_url or iso_urls must be specified"))
		return
	}

	if c.ISOChecksumType == "" {
		errs = append(
			errs, errors.New("The iso_checksum_type must be specified"))
		return
	}
	c.ISOChecksumType = strings.ToLower(c.ISOChecksumType)

	if c.TargetExtension == "" {
		c.TargetExtension = "iso"
	}
	c.TargetExtension = strings.ToLower(c.TargetExtension)

	// Warnings
	if c.ISOChecksumType == "none" {
		warnings = append(warnings,
			"A checksum type of 'none' was specified. Since ISO files are so big,\n"+
				"a checksum is highly recommended.")
		return warnings, errs
	}

	if c.ISOChecksum != "" {
		return warnings, errs
	}

	if c.ISOChecksumURL == "" {
		errs = append(
			errs, errors.New("Due to large file sizes, an iso_checksum is required"))
		return warnings, errs
	}

	for i := range c.ISOUrls {
		u := c.ISOUrls[i]
		nu, err := url.Parse(u)
		if err != nil {
			errs = append(
				errs, fmt.Errorf("Unable to parse %s: %v", u, err))
			continue
		}
		// add checksum in url so that
		// go-getter  to run checksumming for us
		q := nu.Query()
		q.Set("checksum", c.ISOChecksumType+":"+c.ISOChecksum)
		nu.RawQuery = q.Encode()
		c.ISOUrls[i] = nu.String()
	}

	return warnings, errs
}
