//go:build unit || ALL

// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package govcd

import (
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"regexp"
	"testing"
)

// testVcdMockAuthToken is the expected vcdCli.Client.VCDToken value after `Authentication()`
// function passes mock SAML authentication process
// #nosec G101 -- These credentials are fake for testing purposes
const testVcdMockAuthToken = "e3b02b30b8ff4e87ac38db785b0172b5"

// #nosec G101 -- These credentials are fake for testing purposes
const testVcdMockAuthTokenBearer = "eyJhbGciOiJSUzI1NiK1.eyJpc3MiOiJhOTNjOWRiOS03NDcxLTMxOTItOGQwOS1hOGY3ZWVkYTg1ZjlANTU5YmU3OTQtMDE2Yy00NjA3LWE3MmEtNGFiNDdlZTJhYjAwIiwic3ViIjoiYWRtaW5pc3RyYXRvciIsImV4cCI6MTcyNjg5OTcwMiwidmVyc2lvbiI6InZjbG91ZF8xLjAiLCJqdGkiOiI2ZDc2NTNkNTlkMGE0NzVmOTc1Y2M5MTViMTBlY2Q3YyJ9.SEHKnYs-x245KKeyGfaM4PMqUC1lMJie8d_xFn6Qilwr1eEteOsGSj0QB5ee6VPx5wACC1XUf9hqADSV-PQpI_J0u9Z9GZ5bmlN-UJIhuJzmaUEevjCV7z45Z9UewPQZXyMMNOrZiAe6lH_g9ESYJCzoP0YgV4fg5GzkNflZRTpCrLwRNmc54w09TWzmC7Xhoyyh308QjFwdvTAxEUD6yJ7nABEzf65ETXIzYb9fS-H9ZN81x1V1gxr1F-VQXarWoLT85uYcke0KrV19ysE6hwbtnNb15X2oBgt5TWkRF4cCu-MwGqh9T2p3KdxHW9aC-7FSM-vr9SGOx5ojhlZBcw"

// samlMockServer struct allows to attach HTTP handlers to use additional variables (like
// *testing.T) inside those handlers
type samlMockServer struct {
	t *testing.T
}

// TestSamlAdfsAuthenticate is a unit test using mock vCD and ADFS server endpoint to follow
// complete SAML auth flow. The `testVcdMockAuthToken` is expected as an outcome token because
// mock servers return static responses.
//
// Note. A test using real infrastructure is defined in `saml_auth_test.go`
func TestSamlAdfsAuthenticate(t *testing.T) {
	// Spawn mock ADFS server
	adfsServer := testSpawnAdfsServer(t)
	adfsServerHost := adfsServer.URL
	defer adfsServer.Close()

	// Spawn mock vCD instance just enough to cover login details
	vcdServer := spawnVcdServer(t, adfsServerHost, "my-org", "")
	vcdServerHost := vcdServer.URL
	defer vcdServer.Close()

	// Setup vCD client pointing to mock API
	vcdUrl, err := url.Parse(vcdServerHost + "/api")
	if err != nil {
		t.Errorf("got errors: %s", err)
	}
	vcdCli := NewVCDClient(*vcdUrl, true, WithSamlAdfs(true, ""))
	err = vcdCli.Authenticate("fakeUser", "fakePass", "my-org")
	if err != nil {
		t.Errorf("got errors: %s", err)
	}

	// After authentication
	if !vcdCli.Client.UsingBearerToken {
		t.Errorf("expected bearer token")
	}
	if vcdCli.Client.VCDToken != testVcdMockAuthTokenBearer {
		t.Errorf("received token does not match specified one")
	}
}

func TestSamlAdfsAuthenticateWithCookie(t *testing.T) {
	// Spawn mock ADFS server
	adfsServer := testSpawnAdfsServer(t)
	adfsServerHost := adfsServer.URL
	defer adfsServer.Close()

	// Spawn mock vCD instance just enough to cover login details
	vcdServer := spawnVcdServer(t, adfsServerHost, "my-org", "sso-preferred=yes; sso_redirect_org=my-org")
	vcdServerHost := vcdServer.URL
	defer vcdServer.Close()

	// Setup vCD client pointing to mock API
	vcdUrl, err := url.Parse(vcdServerHost + "/api")
	if err != nil {
		t.Errorf("got errors: %s", err)
	}
	vcdCli := NewVCDClient(*vcdUrl, true, WithSamlAdfsAndCookie(true, "", "sso-preferred=yes; sso_redirect_org={{.Org}}"))
	err = vcdCli.Authenticate("fakeUser", "fakePass", "my-org")
	if err != nil {
		t.Errorf("got errors: %s", err)
	}

	// After authentication
	if !vcdCli.Client.UsingBearerToken {
		t.Errorf("expected bearer token")
	}
	if vcdCli.Client.VCDToken != testVcdMockAuthTokenBearer {
		t.Errorf("received token does not match specified one")
	}
}

// spawnVcdServer establishes a mock vCD server with endpoints required to satisfy authentication
func spawnVcdServer(t *testing.T, adfsServerHost, org, expectCookie string) *httptest.Server {
	mockServer := samlMockServer{t}
	mux := http.NewServeMux()
	mux.HandleFunc("/cloud/org/"+org+"/saml/metadata/alias/vcd", mockServer.vCDSamlMetadataHandler)
	mux.HandleFunc("/login/"+org+"/saml/login/alias/vcd", mockServer.getVcdAdfsRedirectHandler(adfsServerHost, expectCookie))
	mux.HandleFunc("/api/sessions", mockServer.vCDLoginHandler)
	mux.HandleFunc("/api/versions", mockServer.vCDApiVersionHandler)
	mux.HandleFunc("/api/org", mockServer.vCDApiOrgHandler)

	server := httptest.NewTLSServer(mux)
	if os.Getenv("GOVCD_DEBUG") != "" {
		log.Printf("vCD mock server now listening on %s...\n", server.URL)
	}
	return server
}

// vcdLoginHandler serves mock "/api/sessions"
func (mockServer *samlMockServer) vCDLoginHandler(w http.ResponseWriter, r *http.Request) {
	// We expect POST method and not anything else
	if r.Method != http.MethodPost {
		w.WriteHeader(500)
		return
	}

	expectedHeader := goldenString(mockServer.t, "REQ_api_sessions", "", false)
	if r.Header.Get("Authorization") != expectedHeader {
		w.WriteHeader(500)
		return
	}

	headers := w.Header()
	headers.Add("X-Vcloud-Authorization", testVcdMockAuthToken)
	headers.Add("X-Vmware-Vcloud-Access-Token", testVcdMockAuthTokenBearer)
	resp := goldenBytes(mockServer.t, "RESP_api_sessions", []byte{}, false)
	_, err := w.Write(resp)
	if err != nil {
		panic(err)
	}
}

// vCDApiVersionHandler server mock "/api/versions"
func (mockServer *samlMockServer) vCDApiVersionHandler(w http.ResponseWriter, r *http.Request) {
	// We expect GET method and not anything else
	if r.Method != http.MethodGet {
		w.WriteHeader(500)
		return
	}

	resp := goldenBytes(mockServer.t, "RESP_api_versions", []byte{}, false)
	_, err := w.Write(resp)
	if err != nil {
		panic(err)
	}
}

// vCDApiOrgHandler serves mock "/api/org"
func (mockServer *samlMockServer) vCDApiOrgHandler(w http.ResponseWriter, r *http.Request) {
	// We expect GET method and not anything else
	if r.Method != http.MethodGet {
		w.WriteHeader(500)
		return
	}

	resp := goldenBytes(mockServer.t, "RESP_api_org", []byte{}, false)
	_, err := w.Write(resp)
	if err != nil {
		panic(err)
	}
}

// vCDSamlMetadataHandler serves mock "/cloud/org/" + org + "/saml/metadata/alias/vcd"
func (mockServer *samlMockServer) vCDSamlMetadataHandler(w http.ResponseWriter, r *http.Request) {
	re := goldenBytes(mockServer.t, "RESP_cloud_org_my-org_saml_metadata_alias_vcd", []byte{}, false)
	_, _ = w.Write(re)
}
func (mockServer *samlMockServer) getVcdAdfsRedirectHandler(adfsServerHost, expectCookie string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(500)
			return
		}

		if expectCookie != "" {
			cookie := r.Header.Get("Cookie")
			if cookie != expectCookie {
				w.WriteHeader(500)
			}
		}

		headers := w.Header()
		locationHeaderPayload := goldenString(mockServer.t, "RESP_HEADER_login_my-org_saml_login_alias_vcd", "", false)
		headers.Add("Location", adfsServerHost+locationHeaderPayload)

		w.WriteHeader(http.StatusFound)
	}
}

// testSpawnAdfsServer spawns mock HTTPS server to server ADFS auth endpoint
// "/adfs/services/trust/13/usernamemixed"
func testSpawnAdfsServer(t *testing.T) *httptest.Server {
	mockServer := samlMockServer{t}
	mux := http.NewServeMux()
	mux.HandleFunc("/adfs/services/trust/13/usernamemixed", mockServer.adfsSamlAuthHandler)
	server := httptest.NewTLSServer(mux)
	if os.Getenv("GOVCD_DEBUG") != "" {
		log.Printf("ADFS mock server now listening on %s...\n", server.URL)
	}
	return server
}

// adfsSamlAuthHandler checks that POST request with expected payload is sent and serves response
// sample ADFS response
func (mockServer *samlMockServer) adfsSamlAuthHandler(w http.ResponseWriter, r *http.Request) {
	// it must be POST method and not anything else
	if r.Method != http.MethodPost {
		w.WriteHeader(500)
		return
	}

	// Replace known dynamic strings to 'REPLACED' string
	gotBody, _ := io.ReadAll(r.Body)
	gotBodyString := string(gotBody)
	re := regexp.MustCompile(`(<a:To s:mustUnderstand="1">).*(</a:To>)`)
	gotBodyString = re.ReplaceAllString(gotBodyString, `${1}REPLACED${2}`)

	re2 := regexp.MustCompile(`(<u:Created>).*(</u:Created>)`)
	gotBodyString = re2.ReplaceAllString(gotBodyString, `${1}REPLACED${2}`)

	re3 := regexp.MustCompile(`(<u:Expires>).*(</u:Expires>)`)
	gotBodyString = re3.ReplaceAllString(gotBodyString, `${1}REPLACED${2}`)

	expectedBody := goldenString(mockServer.t, "REQ_adfs_services_trust_13_usernamemixed", gotBodyString, false)
	if gotBodyString != expectedBody {
		w.WriteHeader(500)
		return
	}

	resp := goldenBytes(mockServer.t, "RESP_adfs_services_trust_13_usernamemixed", []byte(""), false)
	_, err := w.Write(resp)
	if err != nil {
		panic(err)
	}
}
