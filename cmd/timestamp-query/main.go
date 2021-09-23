package main

import (
	"bytes"
	"crypto"
	"encoding/base64"
	"encoding/json"
	"flag"
	"github.com/bingoohuang/jj"
	"github.com/digitorus/timestamp"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
)

func main() {
	request := envFlag("request", "ExampleCreateRequestParseResponse", "timestamp request")
	server := envFlag("url", "https://freetsa.org/tsr", "timestamp server address")
	user := envFlag("user", "", "username:password")
	dumpBody := flag.Bool("dump-body", false, "dump body")
	raw := flag.Bool("raw", false, "keep json raw output (no pretty)")
	flag.Parse()

	log.Printf("Request: %s", *request)
	log.Printf("URL: %s", *server)
	log.Printf("User: %s", *user)
	log.Printf("Dump-Body: %v", *dumpBody)

	if strings.HasPrefix(*server, "@") {
		respBody, err := ioutil.ReadFile((*server)[1:])
		logFatal(err)
		parseResponse(respBody, *raw, *dumpBody)
		return
	}

	opts := &timestamp.RequestOptions{Hash: crypto.SHA256, Certificates: true}
	tsq, err := timestamp.CreateRequest(strings.NewReader(*request), opts)
	logFatal(err)

	r, err := http.NewRequest("POST", *server, bytes.NewReader(tsq))
	logFatal(err)

	r.Header.Set("Content-Type", "application/timestamp-query")
	if *user != "" {
		r.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(*user)))
	}

	rDump, _ := httputil.DumpRequest(r, *dumpBody)
	log.Printf("Request:\n%s", rDump)

	tsr, err := http.DefaultClient.Do(r)
	logFatal(err)

	respDump, _ := httputil.DumpResponse(tsr, *dumpBody)
	log.Printf("Response:\n%s", respDump)

	resp, err := ioutil.ReadAll(tsr.Body)
	logFatal(err)

	parseResponse(resp, *raw, *dumpBody)
}

func parseResponse(resp []byte, raw, dumpBody bool) {
	if v, err := timestamp.ParseResponse(resp); err == nil {
		j := jsonify(v)
		if !raw {
			j = jj.Pretty(j)
		}
		log.Printf("Resp: %s", j)
	} else if !dumpBody {
		log.Printf("Resp: %s", resp)
	}
}

func envFlag(name string, value string, usage string) *string {
	return flag.String(name, getEnv(name, value), usage+`, env `+envName(name))
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
	if v := os.Getenv(envName(name)); v != "" {
		return v
	}
	return defaultValue
}

func envName(name string) string {
	return EnvPrefix + strings.ToUpper(name)
}
