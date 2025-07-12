package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/subosito/gotenv"
	"golang.org/x/crypto/acme/autocert"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Domains map[string]string `yaml:"domains"`
}

func init() {
	_ = gotenv.Load()
}

func main() {
	data, err := os.ReadFile("./servers.yml")
	if err != nil {
		log.Fatal(err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		log.Fatal(err)
	}

	fmt.Println(cfg.Domains)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		targetURL := cfg.Domains[r.Host]
		if targetURL == "" {
			http.Error(w, "Unknown host: "+r.Host, http.StatusBadGateway)
			return
		}

		target, err := url.Parse(targetURL)
		if err != nil {
			http.Error(w, "Invalid target URL", http.StatusInternalServerError)
			return
		}

		proxy := httputil.NewSingleHostReverseProxy(target)

		originalDirector := proxy.Director
		proxy.Director = func(req *http.Request) {
			originalDirector(req)
			req.Host = r.Host
		}

		proxy.ServeHTTP(w, r)
	})

	var hostList []string
	for k := range cfg.Domains {
		hostList = append(hostList, k)
	}

	m := &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(hostList...),
		Cache:      autocert.DirCache("certs"),
	}

	log.Println("Starting HTTPS server with autocert on port 443")
	err = http.Serve(m.Listener(), handler)
	if err != nil {
		log.Fatal("HTTPS server error:", err)
	}
}
