package main

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

var ServerHost string = "localhost"
var ServerPort int = 8008

type Client struct {
	defaultClient *http.Client
	customClient  *http.Client
}

func NewClient() *Client {
	tr := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		Dial: (&net.Dialer{
			Timeout: 1 * time.Second,
		}).Dial,
		TLSHandshakeTimeout:   1 * time.Second,
		MaxIdleConnsPerHost:   10,
		DisableKeepAlives:     true,
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
		ResponseHeaderTimeout: 3 * time.Second,
	}
	return &Client{&http.Client{}, &http.Client{Transport: tr}}
}

type AsyncResponse struct {
	resp *http.Response
	err  error
}

func (c *Client) PrintResult(s time.Time, r *http.Response, err error) {
	if err != nil {
		fmt.Printf(">>>>>> error: %v\n", err)
	} else {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			fmt.Printf(">>>>>> error: %v\n", err)
		} else {
			fmt.Printf(">>>>>> result(%v): %v\n", r.StatusCode, string(body))
		}
	}
	ss := time.Now()
	fmt.Printf(">>>>>> use time: %v\n", ss.Sub(s))
}

func (c *Client) ImmediateReturn() {
	fmt.Printf("\nImmediate Return:\n")
	url := fmt.Sprintf("http://%s:%d/immediate_return", ServerHost, ServerPort)
	fmt.Printf("Default Client:\n")
	s := time.Now()
	r, err := c.defaultClient.Get(url)
	c.PrintResult(s, r, err)
	fmt.Printf("Custom Client:\n")
	s = time.Now()
	r, err = c.customClient.Get(url)
	c.PrintResult(s, r, err)
}

func (c *Client) NeverReturn() {
	fmt.Printf("\nNever Return:\n")
	url := fmt.Sprintf("http://%s:%d/never_return", ServerHost, ServerPort)
	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		fmt.Printf("make request instance error: %v\n", err)
	}
	ch := make(chan AsyncResponse)
	cancel := make(chan struct{})
	req.Cancel = cancel
	cancelFlag := false
	fmt.Printf("Default Client:\n")
	s := time.Now()
	go func() {
		r, err := c.defaultClient.Do(req)
		ch <- AsyncResponse{r, err}
	}()
	select {
	case <-time.After(15 * time.Second):
		close(cancel)
		cancelFlag = true
		fmt.Printf("Not return, use %v\n", time.Now().Sub(s))
	case ar := <-ch:
		c.PrintResult(s, ar.resp, ar.err)
	}

	if cancelFlag {
		a := <-ch
		c.PrintResult(s, a.resp, a.err)
	}

	fmt.Printf("Custom Client:\n")
	s = time.Now()
	r, err := c.customClient.Get(url)
	c.PrintResult(s, r, err)
}

func (c *Client) NeverReturn2() {
	fmt.Printf("\nNever Return2:\n")
	url := fmt.Sprintf("http://%s:%d/never_return", ServerHost, ServerPort)
	fmt.Printf("Default Client:\n")
	c.defaultClient.Timeout = 3 * time.Second
	s := time.Now()
	r, err := c.defaultClient.Get(url)
	c.PrintResult(s, r, err)

	fmt.Printf("Custom Client:\n")
	s = time.Now()
	r, err = c.customClient.Get(url)
	c.PrintResult(s, r, err)
}

func (c *Client) FiveSecondsReturn() {
	fmt.Printf("\n5 Seconds Return:\n")
	url := fmt.Sprintf("http://%s:%d/5_seconds_return", ServerHost, ServerPort)
	fmt.Printf("Default Client:\n")
	s := time.Now()
	r, err := c.defaultClient.Get(url)
	c.PrintResult(s, r, err)
	fmt.Printf("Custom Client:\n")
	s = time.Now()
	r, err = c.customClient.Get(url)
	c.PrintResult(s, r, err)
}

func main() {
	fmt.Printf("Start\n")
	c := NewClient()
	c.ImmediateReturn()
	c.FiveSecondsReturn()
	c.NeverReturn()
	c.NeverReturn2()
	fmt.Printf("End\n")
}
