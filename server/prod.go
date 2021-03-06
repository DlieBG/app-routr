package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"sync"
)

func (fastPort FastPort) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	for hostname, routes := range fastPort {

		if hostname == strings.Split(req.Host, ":")[0] {

			for _, route := range routes {

				if strings.HasPrefix(req.URL.Path, route.Path) {
					req.URL.Path = strings.TrimPrefix(req.URL.Path, route.Path)
					if !strings.HasPrefix(req.URL.Path, "/") {
						req.URL.Path = "/" + req.URL.Path
					}

					url, _ := url.Parse(route.Endpoint)

					proxy := httputil.NewSingleHostReverseProxy(url)
					proxy.ServeHTTP(res, req)

					return
				}
			}

			fmt.Println("404 - Keine Route gedunden")
			res.WriteHeader(404)
			res.Write([]byte("404 - Keine Route gefunden"))
			return
		}
	}

	fmt.Println("404 - Keinen Host gefunden")
	res.WriteHeader(404)
	res.Write([]byte("404 - Keinen Host gefunden"))
}

func scheduleProd(config Config, waiter *sync.WaitGroup) {
	fastCfg := config.toFastCfg()

	fmt.Println("Listening:")

	for portNumber, port := range fastCfg {
		fmt.Println(portNumber, port)
		waiter.Add(1)
		go func(portNumber int, port FastPort) {
			http.ListenAndServe(":"+strconv.FormatInt(int64(portNumber), 10), port)
		}(portNumber, port)
	}

	fmt.Println()
}
