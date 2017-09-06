package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"time"

	"github.com/gobwas/glob"
)

type hostPortPair struct {
	host glob.Glob
	port string
}

func buildProxy(addr string, pMap []hostPortPair, defPort string) *http.Server {
	director := func(req *http.Request) {
		host, _, err := net.SplitHostPort(req.Host)
		if err != nil {
			host = req.Host
		}
		port := defPort
		for _, pair := range pMap {
			if pair.host.Match(host) {
				port = pair.port
				break
			}
		}
		req.URL.Scheme = "http"
		req.URL.Host = "localhost:" + port
		log.Printf("%s -> %s", req.Host, port)
	}
	return &http.Server{
		ReadTimeout:  time.Minute,
		WriteTimeout: time.Minute,
		IdleTimeout:  3 * time.Minute,
		Handler:      &httputil.ReverseProxy{Director: director},
		Addr:         addr,
	}
}

func buildMapAndDefPort(mapList []string) ([]hostPortPair, string, error) {
	var defPort = ""
	var pMap = make([]hostPortPair, len(mapList))
	for i, mapping := range mapList {
		host, port, err := net.SplitHostPort(mapping)
		if err != nil {
			return pMap, defPort, err
		}
		pMap[i] = hostPortPair{host: glob.MustCompile(host), port: port}
		if defPort == "" {
			defPort = port
		}
	}
	return pMap, defPort, nil
}

func main() {
	addr := flag.String("addr", ":http", "which address listen to")
	flag.Usage = func() {
		fmt.Printf("Usage: %s [-addr=[iface]:port] domain:port [domain:port ...]", os.Args[0])
	}
	flag.Parse()
	pMap, defPort, err := buildMapAndDefPort(flag.Args())
	if err != nil || len(pMap) == 0 {
		flag.Usage()
		os.Exit(1)
	}
	log.Fatal(buildProxy(*addr, pMap, defPort).ListenAndServe())
}
