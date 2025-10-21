package ext

import (
	"io"
	"net/http"
	"net/url"

	"github.com/woolawin/catalogue/internal"
)

func NewHTTP() *HTTP {
	return &HTTP{client: http.Client{}}
}

type HTTP struct {
	client http.Client
}

func (impl *HTTP) Fetch(url *url.URL) ([]byte, error) {
	request, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, internal.ErrOf(err, "failed to fetch '%s'", url.String())
	}

	response, err := impl.client.Do(request)
	if err != nil {
		return nil, internal.ErrOf(err, "request failed to '%s'", url.String())
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, internal.ErrOf(err, "failed to read response from %s", url.String())
	}
	return body, nil
}
