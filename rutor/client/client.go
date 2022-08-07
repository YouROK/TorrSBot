package client

import (
	"context"
	"errors"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"time"
)

func GetNic(link, referer, cookie string) (string, error) {
	var (
		dnsResolverIP        = "195.10.195.195:53"
		dnsResolverProto     = "udp"
		dnsResolverTimeoutMs = 10000
	)

	dialer := &net.Dialer{
		Resolver: &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				d := net.Dialer{
					Timeout: time.Duration(dnsResolverTimeoutMs) * time.Millisecond,
				}
				return d.DialContext(ctx, dnsResolverProto, dnsResolverIP)
			},
		},
		Timeout:   120 * time.Second,
		KeepAlive: 120 * time.Second,
	}

	dialContext := func(ctx context.Context, network, addr string) (net.Conn, error) {
		return dialer.DialContext(ctx, network, addr)
	}

	transport := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           dialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	httpClient := &http.Client{Transport: transport}

	req, err := http.NewRequest("GET", link, nil)

	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/75.0.3770.100 Safari/537.36")
	if cookie != "" {
		req.Header.Set("cookie", cookie)
	}
	if referer != "" {
		req.Header.Set("referer", referer)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func Get(link, referer, cookie string) (string, error) {
	buf, err := GetBuf(link, referer, cookie)
	if err != nil {
		return "", err
	}
	return string(buf), nil
}

func GetBuf(link, referer, cookie string) ([]byte, error) {
	var httpClient *http.Client
	httpClient = &http.Client{
		Timeout: 120 * time.Second,
	}
	req, err := http.NewRequest("GET", link, nil)

	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/75.0.3770.100 Safari/537.36")
	if cookie != "" {
		req.Header.Set("cookie", cookie)
	}
	if referer != "" {
		req.Header.Set("referer", referer)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Println("Error get link:", link, resp.StatusCode, resp.Status)
		return nil, errors.New(resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
