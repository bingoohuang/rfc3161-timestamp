package main

import (
	"bytes"
	"crypto"
	"encoding/base64"
	"encoding/json"
	"flag"
	"github.com/digitorus/timestamp"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

const EnvPrefix = "TIMESTAMP_QUERY_"

func getEnv(name, defaultValue string) string {
	envValue := os.Getenv(EnvPrefix + strings.ToUpper(name))
	if envValue == "" {
		return defaultValue
	}
	return envValue
}

func main() {
	timestampRequest := flag.String("request",
		getEnv("request", "ExampleCreateRequestParseResponse"), "timestamp request")
	timestampServer := flag.String("url",
		getEnv("url", "https://freetsa.org/tsr"), "timestamp server address")
	user := flag.String("user", getEnv("user", ""), "username:password")
	flag.Parse()

	tsq, err := timestamp.CreateRequest(strings.NewReader(*timestampRequest),
		&timestamp.RequestOptions{Hash: crypto.SHA256, Certificates: true})
	if err != nil {
		log.Fatal(err)
	}

	r, err := http.NewRequest("POST", *timestampServer, bytes.NewReader(tsq))
	if err != nil {
		log.Fatal(err)
	}
	r.Header.Set("Content-Type", "application/timestamp-query")
	if *user != "" {
		basicAuth := base64.StdEncoding.EncodeToString([]byte(*user))
		r.Header.Set("Authorization", "Basic "+basicAuth)
	}

	tsr, err := http.DefaultClient.Do(r)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("StatusCode: %d", tsr.StatusCode)
	log.Printf("Header: %s", jsonify(tsr.Header))

	resp, err := ioutil.ReadAll(tsr.Body)
	if err != nil {
		log.Fatal(err)
	}

	tsResp, err := timestamp.ParseResponse(resp)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Resp: %s", jsonify(tsResp))
}

func jsonify(v interface{}) []byte {
	data, _ := json.Marshal(v)
	return data
}
