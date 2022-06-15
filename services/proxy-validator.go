package services

import (
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type ProxyValidator struct {
	*http.Request
	Timeout int
}

func (pv *ProxyValidator) Validate(proxy *url.URL) (float64, error) {
	start := time.Now()
	tr := &http.Transport{
		Proxy: http.ProxyURL(proxy),
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   time.Second * time.Duration(pv.Timeout),
	}
	resp, err := client.Do(pv.Request)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("status code: %d", resp.StatusCode)
	}
	return time.Since(start).Seconds(), nil
}
