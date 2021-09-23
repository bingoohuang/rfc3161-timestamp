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

func main() {
	request := envFlag("request", "ExampleCreateRequestParseResponse", "timestamp request")
	server := envFlag("url", "https://freetsa.org/tsr", "timestamp server address")
	user := envFlag("user", "", "username:password")
	flag.Parse()

	opts := &timestamp.RequestOptions{Hash: crypto.SHA256, Certificates: true}
	tsq, err := timestamp.CreateRequest(strings.NewReader(*request), opts)
	logFatal(err)

	r, err := http.NewRequest("POST", *server, bytes.NewReader(tsq))
	logFatal(err)

	r.Header.Set("Content-Type", "application/timestamp-query")
	if *user != "" {
		r.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(*user)))
	}

	tsr, err := http.DefaultClient.Do(r)
	logFatal(err)

	log.Printf("StatusCode: %d", tsr.StatusCode)
	log.Printf("Header: %s", jsonify(tsr.Header))

	resp, err := ioutil.ReadAll(tsr.Body)
	logFatal(err)

	tsResp, err := timestamp.ParseResponse(resp)
	if err != nil {
		log.Printf("Resp: %s", resp)
	} else {
		log.Printf("Resp: %s", jsonify(tsResp))
	}
}

func envFlag(name string, value string, usage string) *string {
	return flag.String(name, getEnv(name, value), usage)
}

func logFatal(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func jsonify(v interface{}) []byte {
	data, _ := json.Marshal(v)
	return data
}

const EnvPrefix = "TIMESTAMP_QUERY_"

func getEnv(name, defaultValue string) string {
	if v := os.Getenv(EnvPrefix + strings.ToUpper(name)); v != "" {
		return v
	}
	return defaultValue
}
