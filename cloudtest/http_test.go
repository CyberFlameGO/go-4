//  Copyright 2017 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cloudtest_test

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"testing"

	"net/http/httptest"

	"github.com/m-lab/go/cloudtest"
)

func init() {
	// Always prepend the filename and line number.
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

// Tests the LoggingClient
func TestNewLoggingClientBasic(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stdout)

	// Use a local test server.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, client")
	}))
	defer ts.Close()

	// Use a logging client.
	client := cloudtest.NewLoggingClient()

	// Send request through the client to the test URL.
	_, err := client.Get(ts.URL)
	if err != nil {
		t.Error(err)
	}

	// Check that the log buffer contains the expected output.
	if !strings.Contains(buf.String(), "Request:\n") {
		t.Error("Should contain Request: ", buf.String())
	}
	if !strings.Contains(buf.String(), "Response body:") {
		t.Error("Should contain response body")
	}
	if !strings.Contains(buf.String(), "Hello, client") {
		t.Error("Should contain Hello, client: ", buf.String())
	}
}

// Tests the LoggingClient
func TestLoggingClientBasic(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stdout)

	// Use a local test server.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, client")
	}))
	defer ts.Close()

	// Use a logging client.
	client, err := cloudtest.LoggingClient(&http.Client{})
	if err != nil {
		t.Error(err)
	}

	// Send request through the client to the test URL.
	_, err = client.Get(ts.URL)
	if err != nil {
		t.Error(err)
	}

	// Check that the log buffer contains the expected output.
	if !strings.Contains(buf.String(), "Request:\n") {
		t.Error("Should contain Request: ", buf.String())
	}
	if !strings.Contains(buf.String(), "Response body:") {
		t.Error("Should contain response body")
	}
	if !strings.Contains(buf.String(), "Hello, client") {
		t.Error("Should contain Hello, client: ", buf.String())
	}
}

// Tests the ChannelClient, which pulls responses from a provided
// channel.
func TestChannelClientBasic(t *testing.T) {
	c := make(chan *http.Response, 10)
	client := cloudtest.NewChannelClient(c)

	resp := &http.Response{}
	resp.StatusCode = http.StatusOK
	resp.Status = "OK"
	c <- resp
	resp, err := client.Get("http://foobar")
	log.Printf("%v\n", resp)
	if err != nil {
		t.Error(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Error("Response should be OK: ", resp.Status)
	}
}
