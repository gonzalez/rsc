package rsapi

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"
)

// Log request, dump its content if required then make request and log response and dump it.
func (a *Api) PerformRequest(req *http.Request) (*http.Response, error) {
	var id string
	var startedAt time.Time
	if a.Logger != nil {
		startedAt = time.Now()
		b := make([]byte, 6)
		io.ReadFull(rand.Reader, b)
		id = base64.StdEncoding.EncodeToString(b)
		a.Logger.Printf("[%s] %s %s", id, req.Method, req.URL.String())
	}
	if a.DumpRequestResponse == Debug {
		b, err := httputil.DumpRequestOut(req, true)
		if err == nil {
			fmt.Printf("%s\n", string(b))
		} else {
			fmt.Printf("Failed to dump request content - %s\n", err)
		}
	}
	var reqBody []byte
	if a.DumpRequestResponse == Json {
		var err error
		reqBody, err = ioutil.ReadAll(req.Body)
		if err != nil {
			fmt.Printf("Failed to load request body for dump: %s\n", err)
		}
	}
	// Sign last so auth headers don't get printed or logged
	if err := a.Auth.Sign(req, a.Host, a.AccountId); err != nil {
		return nil, err
	}
	// TBD: Read version from same place as command line tool
	req.Header.Set("User-Agent", UA)
	resp, err := a.Client.Do(req)
	if err != nil {
		return nil, err
	}
	if a.Logger != nil {
		d := time.Since(startedAt)
		a.Logger.Printf("[%s] %s in %s", id, resp.Status, d.String())
	}
	if a.DumpRequestResponse == Debug {
		b, err := httputil.DumpResponse(resp, false)
		if err == nil {
			fmt.Printf("--------\n%s", b)
		} else {
			fmt.Printf("Failed to dump response content - %s\n", err)
		}
	} else if a.DumpRequestResponse == Json {
		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("Failed to load response body for dump: %s\n", err)
		}
		reqHeaders := make(map[string]string, len(req.Header))
		for n, hs := range req.Header {
			reqHeaders[n] = strings.Join(hs, ", ")
		}
		respHeaders := make(map[string]string, len(resp.Header))
		for n, hs := range resp.Header {
			respHeaders[n] = strings.Join(hs, ", ")
		}
		dumped := struct {
			Verb       string
			Uri        string
			ReqHeader  map[string]string
			ReqBody    string
			Status     int
			RespHeader map[string]string
			RespBody   string
		}{
			Verb:       req.Method,
			Uri:        req.URL.String(),
			ReqHeader:  reqHeaders,
			ReqBody:    string(reqBody),
			Status:     resp.StatusCode,
			RespHeader: respHeaders,
			RespBody:   string(respBody),
		}
		b, err := json.MarshalIndent(dumped, "", "    ")
		if err == nil {
			fmt.Printf("%s\n", string(b))
		} else {
			fmt.Printf("Failed to dump request content - %s\n", err)
		}
	}

	return resp, err
}

// Deserialize JSON response into generic object.
// If the response has a "Location" header then the returned object is a map with one key "Location"
// containing the value of the header.
func (a *Api) LoadResponse(resp *http.Response) (interface{}, error) {
	defer resp.Body.Close()
	var respBody interface{}
	jsonResp, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Failed to read response (%s)", err)
	}
	if len(jsonResp) > 0 {
		err = json.Unmarshal(jsonResp, &respBody)
		if err != nil {
			return nil, fmt.Errorf("Failed to load response (%s)", err)
		}
	}
	// Special case for "Location" header, assume that if there is a location there is no body
	loc := resp.Header.Get("Location")
	if len(loc) > 0 {
		var bodyMap = make(map[string]interface{})
		bodyMap["Location"] = loc
		respBody = interface{}(bodyMap)
	}
	return respBody, err
}
