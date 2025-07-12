package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/subosito/gotenv"
	"golang.org/x/crypto/acme/autocert"
	"gopkg.in/yaml.v3"
)

func init() {
	_ = gotenv.Load()
}

type Config struct {
	Domains map[string]string `yaml:"domains"`
}



func main() {

	data, err := os.ReadFile("./servers.yml")
	if err != nil {
		panic(err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		panic(err)
	}
	fmt.Println(cfg.Domains)
	
	
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		targetURL :=  cfg.Domains[r.Host]

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

		// Optionally preserve original Host header
		originalDirector := proxy.Director
		proxy.Director = func(req *http.Request) {
			originalDirector(req)
			req.Host = r.Host
		}

		proxy.ServeHTTP(w, r)
	})

	go func() {
		ginApp := gin.New()

		gin.SetMode(gin.ReleaseMode)
		ginApp.GET("/", func(ctx *gin.Context) {
			ctx.JSON(200, "osas")
		})
		ginApp.Run(":3001")
	}()

	var hostList= []string{}

	for k:=range cfg.Domains{
		hostList = append(hostList, k)
	}

	m := &autocert.Manager{
		Prompt: autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(hostList...),
		Cache: autocert.DirCache("certs"),
	}

	server1 := &http.Server{
		Addr:      ":80",
		Handler:   handler,
	}

	server2 := &http.Server{
		Addr:      ":443",
		TLSConfig: m.TLSConfig(),
		Handler:   handler,
	}

	// log.Fatal(server.ListenAndServeTLS("", ""))
	go log.Fatal(server1.ListenAndServe())
	log.Fatal(server2.ListenAndServeTLS("",""))

}
