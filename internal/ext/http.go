package ext

import (
	"io"
	"net/http"
	"net/url"

	"github.com/woolawin/catalogue/internal"
)

type HTTP interface {
	Fetch(url *url.URL) ([]byte, error)
}

func NewHTTP() HTTP {
	return &httpImpl{client: http.Client{}}
}

type httpImpl struct {
	client http.Client
}

func (impl *httpImpl) Fetch(url *url.URL) ([]byte, error) {
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
