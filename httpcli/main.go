package httpcli

import (
	"log"
	"net/http"
	"net/http/httputil"
)

// loggingTransport wraps around an existing http.RoundTripper, allowing
// you to log requests and responses.
type loggingTransport struct {
	Transport http.RoundTripper
}

// RoundTrip executes a single HTTP transaction and logs the request and response.
func (c *loggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Log the request
	var (
	//ctx = req.Context()
	)

	reqData, err := httputil.DumpRequestOut(req, true)
	if err == nil {
		log.Println("Request:", string(reqData))
	}

	// Use the embedded Transport to execute the actual request
	resp, err := c.Transport.RoundTrip(req)

	// Log the response
	if err == nil {
		respData, err := httputil.DumpResponse(resp, true)
		if err == nil {
			log.Println("Response:", string(respData))
		}
	}

	return resp, err
}

func main() {
	// Create a new http.Client with the custom Transport
	client := &http.Client{
		Transport: &loggingTransport{
			Transport: http.DefaultTransport,
		},
	}

	// Use the client to make requests as usual
	_, err := client.Get("https://www.google.com")
	if err != nil {
		log.Fatal(err)
	}
}
