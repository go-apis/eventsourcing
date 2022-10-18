package identity

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/textproto"
	"net/url"
	"strings"
)

// Hop-by-hop headers. These are removed when sent to the backend.
// As of RFC 7230, hop-by-hop headers are required to appear in the
// Connection header field. These are the headers defined by the
// obsoleted RFC 2616 (section 13.5.1) and are used for backward
// compatibility.
var hopHeaders = []string{
	"Connection",
	"Proxy-Connection", // non-standard but still sent by libcurl and rejected by e.g. google
	"Keep-Alive",
	"Proxy-Authenticate",
	"Proxy-Authorization",
	"Te",      // canonicalized version of "TE"
	"Trailer", // not Trailers per URL above; https://www.rfc-editor.org/errata_search.php?eid=4522
	"Transfer-Encoding",
	"Upgrade",
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

// removeConnectionHeaders removes hop-by-hop headers listed in the "Connection" header of h.
// See RFC 7230, section 6.1
func removeConnectionHeaders(h http.Header) {
	for _, f := range h["Connection"] {
		for _, sf := range strings.Split(f, ",") {
			if sf = textproto.TrimString(sf); sf != "" {
				h.Del(sf)
			}
		}
	}
}

func newRequest(server string, r *http.Request) (*http.Request, error) {
	body := &SessionRequest{
		Method: r.Method,
		Path:   r.URL.String(),
	}
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}
	queryURL, err := serverURL.Parse("/session/authn")
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", queryURL.String(), bodyReader)
	if err != nil {
		return nil, err
	}
	copyHeader(req.Header, r.Header)
	removeConnectionHeaders(req.Header)
	for _, h := range hopHeaders {
		req.Header.Del(h)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-Forwarded-Host", r.Host)

	return req, nil
}

func NewFetcher[T User](server string) Fetch {
	client := &http.Client{}

	return func(r *http.Request) (User, error) {
		req, err := newRequest(server, r)
		if err != nil {
			return nil, err
		}
		req = req.WithContext(r.Context())
		rsp, err := client.Do(req)
		if err != nil {
			return nil, err
		}

		bodyBytes, err := ioutil.ReadAll(rsp.Body)
		if err != nil {
			return nil, err
		}
		defer func() { _ = rsp.Body.Close() }()

		switch rsp.StatusCode {
		case http.StatusOK:
			var dest T
			if err := json.Unmarshal(bodyBytes, &dest); err != nil {
				return nil, err
			}
			return dest, nil
		default:
			var dest ErrorMessage
			if err := json.Unmarshal(bodyBytes, &dest); err != nil {
				return nil, err
			}
			return nil, &dest
		}
	}
}
