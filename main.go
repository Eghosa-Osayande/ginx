package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/acme/autocert"
)

// Map domains to target local servers
var routes = map[string]string{
	"osas.localhost:8000": "http://localhost:3001",
	"tawoh.localhost:8000": "http://localhost:3002",
}

func main() {
	

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		targetURL, ok := routes[r.Host]
		if !ok {
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


	go func() {
		ginApp := gin.New()

		gin.SetMode(gin.ReleaseMode)
		ginApp.GET("/", func(ctx *gin.Context) {
			ctx.JSON(200, "tawoh")
		})
		ginApp.Run(":3002")
	}()

	m := &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(
			"tawoh.com",
			"osadolor.com",
		),
		Cache:      autocert.DirCache("certs"),
	}


	server := &http.Server{
		Addr:      ":8000",
		TLSConfig: m.TLSConfig(),
		Handler:   handler,
	}

	log.Fatal(server.ListenAndServeTLS("", ""))

}
