package fluentd

import (
	"fmt"
	"net/url"
	"strings"

	log "github.com/sirupsen/logrus"
)

// LokiURL returns the endpoint url without any further processing
func (conf *outputLabelConf) LokiURL() string {
	url, err := url.Parse(conf.Target.URL)
	if err != nil {
		log.Errorf("Failed parsing endpoint as a valid URL")
		return ""
	}

	return fmt.Sprintf("%s://%s", url.Scheme, url.Host)
}

// LokiTenant returns the loki tenant ID
func (conf *outputLabelConf) LokiTenant() string {
	if conf.Target.Loki != nil {
		if conf.Target.Loki.TenantID != "" {
			return conf.Target.Loki.TenantID
		}
	}

	url, err := url.Parse(conf.Target.URL)
	if err != nil {
		log.Errorf("Failed to extract Loki tenant from endpoint url %q: %s", conf.Target.URL, err)
		return ""
	}
	return strings.Trim(url.Path, "/")
}
