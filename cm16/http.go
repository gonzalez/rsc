package cm16 // import "gopkg.in/rightscale/rsc.v1/cm16"

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"gopkg.in/rightscale/rsc.v1/rsapi"
)

// BuildRequest builds a HTTP request from a resource name and href and an action name and
// parameters.
// It is intended for generic clients that need to consume APIs in a generic maner.
// The method builds an HTTP request that can be fed to PerformRequest.
func (a *Api) BuildRequest(resource, action, href string, params rsapi.ApiParams) (*http.Request, error) {
	// First lookup metadata
	res, ok := GenMetadata[resource]
	if !ok {
		return nil, fmt.Errorf("No resource with name '%s'", resource)
	}
	act := res.GetAction(action)
	if act == nil {
		return nil, fmt.Errorf("No action with name '%s' on %s", action, resource)
	}

	// Now lookup action request HTTP method, url, params and payload.
	vars, err := res.ExtractVariables(href)
	if err != nil {
		return nil, err
	}
	actionUrl, err := act.Url(vars)
	if err != nil {
		return nil, err
	}
	queryParams := make(rsapi.ApiParams, len(act.QueryParamNames))
	for _, n := range act.QueryParamNames {
		queryParams[n] = params[n]
	}
	return a.buildHttpRequest(actionUrl.Path, queryParams)
}

// Helper function that signs, makes and logs HTTP request
func (a *Api) Dispatch(verb, uri string, params, payload rsapi.ApiParams) (*http.Response, error) {
	req, err := a.buildHttpRequest(uri, params)
	if err != nil {
		return nil, err
	}
	return a.PerformRequest(req)
}

// Helper function that puts together and HTTP request from its uri and params.
func (a *Api) buildHttpRequest(uri string, params rsapi.ApiParams) (*http.Request, error) {
	u := url.URL{
		Host: a.Host,
		Path: uri,
	}
	if params != nil {
		values := u.Query()
		for n, p := range params {
			switch t := p.(type) {
			case string:
				values.Set(n, t)
			case int:
				values.Set(n, strconv.Itoa(t))
			case bool:
				values.Set(n, strconv.FormatBool(t))
			case []string:
				for _, e := range t {
					values.Add(n, e)
				}
			case []int:
				for _, e := range t {
					values.Add(n, strconv.Itoa(e))
				}
			case []bool:
				for _, e := range t {
					values.Add(n, strconv.FormatBool(e))
				}
			case []interface{}:
				for _, e := range t {
					values.Add(n, fmt.Sprintf("%v", e))
				}
			case map[string]string:
				for pn, e := range t {
					values.Add(fmt.Sprintf("%s[%s]", n, pn), e)
				}
			default:
				return nil, fmt.Errorf("Invalid param value <%+v>, value must be a string, an integer, a bool, an array of these types of a map of strings", p)
			}
		}
		u.RawQuery = values.Encode()
	}
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-API-Version", "1.6")
	if a.AccountId > 0 {
		req.Header.Set("X-Account", strconv.Itoa(a.AccountId))
	}
	return req, nil
}
