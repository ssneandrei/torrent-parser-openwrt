package main

import (
 "fmt"
 "io"
 "net/http"
 "net/http/cookiejar"
 "net/url"
 "strings"
)

const maxResponseBody = 8 << 20 // 8 MiB

func newSessionClient(base *http.Client) (*http.Client, error) {
 jar, err:= cookiejar.New(nil)
 if err!= nil {
 return nil, fmt.Errorf("cookie jar: %w", err)
 }

 client:= *base
 client.Jar = jar

 return &client, nil
}

func trackerRequest(
 client *http.Client,
 method string,
 rawURL string,
 form url.Values,
) (*http.Response, error) {
 var body io.Reader = http.NoBody

 if form!= nil {
 body = strings.NewReader(form.Encode())
 }

 req, err:= http.NewRequest(method, rawURL, body)
 if err!= nil {
 return nil, err
 }

 req.Header.Set("User-Agent", defaultUA)
 req.Header.Set("Accept", "text/html,application/xhtml+xml,*/*;q=0.8")
 req.Header.Set("Accept-Language", "ru-RU,ru;q=0.9,en;q=0.5")

 if form!= nil {
 req.Header.Set(
 "Content-Type",
 "application/x-www-form-urlencoded",
 )
 }

 resp, err:= client.Do(req)
 if err!= nil {
 return nil, err
 }

 if resp.StatusCode < 200 || resp.StatusCode >= 400 {
 data, _:= readResponseBody(resp)

 text:= strings.TrimSpace(string(data))
 if len(text) > 300 {
 text = text[:300]
 }

 return nil, fmt.Errorf(
 "HTTP %d %s: %s",
 resp.StatusCode,
 resp.Status,
 text,
 )
 }

 return resp, nil
}

func readResponseBody(resp *http.Response) ([]byte, error) {
 defer resp.Body.Close()

 reader:= io.LimitReader(resp.Body, maxResponseBody+1)

 data, err:= io.ReadAll(reader)
 if err!= nil {
 return nil, err
 }

 if len(data) > maxResponseBody {
 return nil, fmt.Errorf(
 "response is larger than %d bytes",
 maxResponseBody,
 )
 }

 return data, nil
}

func absoluteURL(baseURL, href string) string {
 href = strings.TrimSpace(href)
 if href == "" {
 return ""
 }

 ref, err:= url.Parse(href)
 if err!= nil {
 return ""
 }

 if ref.IsAbs() {
 return ref.String()
 }

 base, err:= url.Parse(strings.TrimRight(baseURL, "/") + "/")
 if err!= nil {
 return ""
 }

 return base.ResolveReference(ref).String()
}
