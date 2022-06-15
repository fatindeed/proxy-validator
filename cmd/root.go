package cmd

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/fatindeed/proxy-validator/services"
	"github.com/rodaine/table"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	// define flags
	rootCmd.Flags().String("file", "proxy.txt", "Proxy file")
	rootCmd.Flags().String("url", "https://www.google.com/", "URL to work with")
	rootCmd.Flags().StringArray("header", nil, "Pass custom header(s) to server")
	rootCmd.Flags().StringArray("cookie", nil, "Send cookies from string")
	rootCmd.Flags().Int("timeout", 10, "Maximum time allowed for connection")
	rootCmd.Flags().Int("concurrency", 10, "Concurrent requests")
	// bind flags
	viper.BindPFlag("file", rootCmd.Flags().Lookup("file"))
	viper.BindPFlag("url", rootCmd.Flags().Lookup("url"))
	viper.BindPFlag("header", rootCmd.Flags().Lookup("header"))
	viper.BindPFlag("cookie", rootCmd.Flags().Lookup("cookie"))
	viper.BindPFlag("timeout", rootCmd.Flags().Lookup("timeout"))
	viper.BindPFlag("concurrency", rootCmd.Flags().Lookup("concurrency"))
	// bind environment variables
	viper.SetEnvPrefix("PROXY_VALIDATOR")
	viper.AutomaticEnv()
}

var (
	// Version is the version of the CLI injected in compilation time
	Version = "dev"
	rootCmd = &cobra.Command{
		Use:     "proxy-validator",
		Version: fmt.Sprintf("1.0.0, build %s", Version),
		Short:   "Validate proxy with specific url",
		RunE: func(cmd *cobra.Command, args []string) error {
			request, err := getRequest()
			if err != nil {
				return err
			}
			pv := services.ProxyValidator{
				Request: request,
				Timeout: viper.GetInt("timeout"),
			}
			proxyList, err := services.GetProxyList(viper.GetString("file"))
			if err != nil {
				return err
			}

			wg := sync.WaitGroup{}
			tbl := table.New("Proxy", "Response")
			queue := make(chan struct{}, viper.GetInt("concurrency"))
			for _, proxy := range proxyList {
				wg.Add(1)
				go func(proxy *url.URL) {
					defer wg.Done()
					queue <- struct{}{}

					elapsed, _ := pv.Validate(proxy)
					if elapsed > 0 {
						tbl.AddRow(proxy, elapsed)
					}
					<-queue
				}(proxy)
			}
			wg.Wait()
			tbl.Print()
			return nil
		},
		SilenceUsage: true,
	}
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func getRequest() (*http.Request, error) {
	testURL := viper.GetString("url")
	u, err := url.Parse(testURL)
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequest("GET", testURL, nil)
	if err != nil {
		return nil, err
	}
	headers := viper.GetStringSlice("header")
	for _, header := range headers {
		pos := strings.Index(header, ":")
		if pos == -1 {
			return nil, fmt.Errorf("invalid header: %s", header)
		}
		request.Header.Set(header[0:pos], strings.Trim(header[pos:], " "))
	}
	cookies := viper.GetStringSlice("cookie")
	for _, cookie := range cookies {
		pos := strings.Index(cookie, "=")
		if pos == -1 {
			return nil, fmt.Errorf("invalid cookie: %s", cookie)
		}
		request.AddCookie(&http.Cookie{
			Name:   cookie[0:pos],
			Value:  url.QueryEscape(cookie[pos:]),
			Domain: u.Host,
			Path:   "/",
		})
	}
	return request, nil
}
