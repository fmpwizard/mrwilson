//Mr Wilson is a service that uses the wit.ai api to understand your text messages and stores them for later retrieval
//It uses nexmo as an sms gateway
package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/user"
	"path/filepath"

	"golang.org/x/crypto/acme/autocert"
)

var config MrWilsonConfig
var db string
var token string

func main() {
	flag.Parse()
	readMrWilsonConfig()
	log.Println("mode is ", mode)
	smsToken()
	db = initDB()
	http.HandleFunc("/status", statusHandler)
	http.Handle("/sms", checkNexmoIP(NexmoHandler))
	http.Handle("/db", checkToken(CSVHandler))
	http.Handle("/recommend", checkToken(RecommendHandler))
	if mode == production {
		m := autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(config.HostNames...),
			Cache:      autocert.DirCache(config.LECacheFilePath),
			Email:      config.LEEmail,
		}
		s := &http.Server{
			Addr:      ":https",
			TLSConfig: &tls.Config{GetCertificate: m.GetCertificate},
		}
		//Listen on port 80 and redirect them all to 443
		go func() {
			log.Fatal(http.ListenAndServe(":80", http.HandlerFunc(httpToHTTPS)))
		}()
		log.Println("Running on port: 443")
		s.ListenAndServeTLS("", "")
	} else {
		log.Println("Running on port: ", config.HTTPPort)
		http.ListenAndServe(fmt.Sprintf(":%v", config.HTTPPort), nil)
	}
}

func smsToken() {
	t, err := createToken()
	if err != nil {
		log.Printf("error generating token, won't be able to login. %s\n", err)
	} else {
		token = t
		if mode == production {
			log.Println("Sending token to HQ")
			sendSMS(config.TokenTo, t)
		} else {
			log.Println("login token is: ", token)
		}
	}
}

//httpToHTTPS redirects all http to https
func httpToHTTPS(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "https://"+r.Host+r.URL.String(), http.StatusMovedPermanently)
}

func initDB() string {
	return config.DBPath + "/mr.wilson.csv"
}

func readMrWilsonConfig() {
	ReadProps()
	log.Printf("wit.ai bearer length: %d\ndb path: %s\ncert path: %s", len(config.WitAccessToken), config.DBPath, config.LECacheFilePath)
}

//MrWilsonConfig hold the configuration for Cortex to work.
type MrWilsonConfig struct {
	HTTPPort        int32
	WitAccessToken  string
	DBPath          string
	NexmoAPIKey     string
	NexmoAPISecret  string
	NexmoFromNumber string
	HostNames       []string
	LECacheFilePath string
	LEEmail         string
	TokenTo         string
}

//ReadProps reads the .rackness.json file from your home dir and fill in the settings info
func ReadProps() {
	if u, err := user.Current(); err == nil {
		path := filepath.Join(u.HomeDir, ".mr.wilson.json")
		if _, err := os.Stat(path); os.IsNotExist(err) {
			config.HTTPPort = 1212
			config.WitAccessToken = "enter wit access token here."
			config.DBPath = "/path/to/csv/file/"
			sampleConfig, _ := json.MarshalIndent(config, "", "  ")
			ioutil.WriteFile(path, sampleConfig, os.FileMode(int(0400)))
			fmt.Printf("Created a sample config file: %s\nPlease update it with your information.\n", path)
			os.Exit(1)
		}

		content, err := ioutil.ReadFile(path)
		if err != nil {
			log.Fatalf("Error reading props file, got: %v\n", err)
		}

		err = json.Unmarshal(content, &config)
		if err != nil {
			log.Fatalf("Invalid Props file, got: %v\n", err)
		}

		// Strip trailing slashes from DB path.
		for len(config.DBPath) > 0 && os.IsPathSeparator(config.DBPath[len(config.DBPath)-1]) {
			config.DBPath = config.DBPath[0 : len(config.DBPath)-1]
		}
		// Strip trailing slashes from let's encrypt cache path.
		for len(config.LECacheFilePath) > 0 && os.IsPathSeparator(config.LECacheFilePath[len(config.LECacheFilePath)-1]) {
			config.LECacheFilePath = config.LECacheFilePath[0 : len(config.LECacheFilePath)-1]
		}
		if m := os.Getenv("RUNMODE"); m == "production" {
			mode = production
		}
		log.Printf("Running in %s mode.\n", mode)
	}
}

// RedirectHTTP is an HTTP handler (suitable for use with http.HandleFunc)
// that responds to all requests by redirecting to the same URL served over HTTPS.
// It should only be invoked for requests received over HTTP.
// taken from rsc.io/letsencrypt
func RedirectHTTP(w http.ResponseWriter, r *http.Request) {
	if r.TLS != nil || r.Host == "" {
		http.Error(w, "not found", 404)
	}

	u := r.URL
	u.Host = r.Host
	u.Scheme = "https"
	http.Redirect(w, r, u.String(), 302)
}

// statusHandler is used by our monitoring system
func statusHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
}
