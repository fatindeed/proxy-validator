package services

import (
	"fmt"
	"net/url"
	"os"
	"strings"
)

func GetProxyList(file string) ([]*url.URL, error) {
	contents, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(contents), "\n")
	proxyList := make([]*url.URL, 0)
	for _, line := range lines {
		if line[0:4] != "http" {
			line = "http://" + line
		}
		proxy, err := url.Parse(line)
		if err != nil {
			return nil, err
		}
		proxyList = append(proxyList, proxy)
	}
	if len(proxyList) == 0 {
		return nil, fmt.Errorf("no proxy found in %s", file)
	}
	return proxyList, nil
}
