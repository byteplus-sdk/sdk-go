package core

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/valyala/fasthttp"
)

const (
	timeFormatV4 = "20060102T150405Z"
)

type Credential struct {
	AccessKeyID     string
	SecretAccessKey string
	Region          string
	Service         string
	SessionToken    string
}

type metadata struct {
	algorithm       string
	credentialScope string
	signedHeaders   string
	date            string
	region          string
	service         string
}

var now = func() time.Time {
	return time.Now().UTC()
}

func VolcSign(req *fasthttp.Request, cred Credential) *fasthttp.Request {
	prepareRequestV4(req)

	meta := &metadata{}
	meta.service, meta.region = cred.Service, cred.Region

	// Task 1
	hashedCanonReq := hashedCanonicalRequestV4(req, meta)

	// Task 2
	stringToSignRet := stringToSign(req, hashedCanonReq, meta)

	// Task 3

	signingKeyRet := signingKey(cred.SecretAccessKey, meta.date, meta.region, meta.service)
	signatureRet := signature(signingKeyRet, stringToSignRet)

	req.Header.Set("Authorization", buildAuthHeader(signatureRet, meta, cred))

	if cred.SessionToken != "" {
		req.Header.Set("X-Security-Token", cred.SessionToken)
	}

	return req
}

func prepareRequestV4(req *fasthttp.Request) *fasthttp.Request {
	necessaryDefaults := map[string]string{
		"Content-Type": "application/x-www-form-urlencoded; charset=utf-8",
		"X-Date":       timestampV4(),
	}

	for header, value := range necessaryDefaults {
		if len(req.Header.Peek(header)) == 0 {
			req.Header.Set(header, value)
		}
	}

	if len(req.URI().Path()) == 0 {
		req.URI().SetPath("/")
	}

	return req
}

func timestampV4() string {
	return now().Format(timeFormatV4)
}

func hashedCanonicalRequestV4(req *fasthttp.Request, meta *metadata) string {
	payload := req.Body()
	payloadHash := hashSHA256(payload)
	req.Header.Set("X-Content-Sha256", payloadHash)

	req.Header.Set("Host", string(req.URI().Host()))

	var sortedHeaderKeys []string
	req.Header.VisitAll(func(keyBytes, valueBytes []byte) {
		key := strings.ToLower(string(keyBytes))
		switch key {
		case "content-type", "content-md5", "host":
		default:
			if !strings.HasPrefix(key, "x-") {
				return
			}
		}
		sortedHeaderKeys = append(sortedHeaderKeys, key)
	})
	sort.Strings(sortedHeaderKeys)

	var headersToSign string
	for _, key := range sortedHeaderKeys {
		value := strings.TrimSpace(string(req.Header.Peek(key)))
		if key == "host" {
			if strings.Contains(value, ":") {
				split := strings.Split(value, ":")
				port := split[1]
				if port == "80" || port == "443" {
					value = split[0]
				}
			}
		}
		headersToSign += key + ":" + value + "\n"
	}
	meta.signedHeaders = concat(";", sortedHeaderKeys...)

	// keep k,v order with server
	urlQuery := url.Values{}
	req.URI().QueryArgs().VisitAll(func(key, value []byte) {
		urlQuery.Add(string(key), string(value))
	})
	canonicalRequest := concat("\n", string(req.Header.Method()), normuri(string(req.URI().Path())), normquery(urlQuery.Encode()), headersToSign, meta.signedHeaders, payloadHash)

	return hashSHA256([]byte(canonicalRequest))
}

func hashSHA256(content []byte) string {
	h := sha256.New()
	h.Write(content)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func readAndReplaceBody(request *http.Request) []byte {
	if request.Body == nil {
		return []byte{}
	}
	payload, _ := ioutil.ReadAll(request.Body)
	request.Body = ioutil.NopCloser(bytes.NewReader(payload))
	return payload
}

func concat(delim string, str ...string) string {
	return strings.Join(str, delim)
}

func normuri(uri string) string {
	parts := strings.Split(uri, "/")
	for i := range parts {
		parts[i] = encodePathFrag(parts[i])
	}
	return strings.Join(parts, "/")
}

func encodePathFrag(s string) string {
	hexCount := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		if shouldEscape(c) {
			hexCount++
		}
	}
	t := make([]byte, len(s)+2*hexCount)
	j := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		if shouldEscape(c) {
			t[j] = '%'
			t[j+1] = "0123456789ABCDEF"[c>>4]
			t[j+2] = "0123456789ABCDEF"[c&15]
			j += 3
		} else {
			t[j] = c
			j++
		}
	}
	return string(t)
}

func shouldEscape(c byte) bool {
	if 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z' {
		return false
	}
	if '0' <= c && c <= '9' {
		return false
	}
	if c == '-' || c == '_' || c == '.' || c == '~' {
		return false
	}
	return true
}

func normquery(queryString string) string {
	return strings.Replace(queryString, "+", "%20", -1)
}

func stringToSign(req *fasthttp.Request, hashedCanonReq string, meta *metadata) string {
	requestTs := string(req.Header.Peek("X-Date"))

	meta.algorithm = "HMAC-SHA256"
	meta.date = tsDate(requestTs)
	meta.credentialScope = concat("/", meta.date, meta.region, meta.service, "request")

	return concat("\n", meta.algorithm, requestTs, meta.credentialScope, hashedCanonReq)
}

func tsDate(timestamp string) string {
	return timestamp[:8]
}

func signingKey(secretKey, date, region, service string) []byte {
	kDate := hmacSHA256([]byte(secretKey), date)
	kRegion := hmacSHA256(kDate, region)
	kService := hmacSHA256(kRegion, service)
	kSigning := hmacSHA256(kService, "request")
	return kSigning
}

func signature(signingKey []byte, stringToSign string) string {
	return hex.EncodeToString(hmacSHA256(signingKey, stringToSign))
}

func hmacSHA256(key []byte, content string) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(content))
	return mac.Sum(nil)
}

func buildAuthHeader(signature string, meta *metadata, keys Credential) string {
	credential := keys.AccessKeyID + "/" + meta.credentialScope

	return meta.algorithm +
		" Credential=" + credential +
		", SignedHeaders=" + meta.signedHeaders +
		", Signature=" + signature
}
