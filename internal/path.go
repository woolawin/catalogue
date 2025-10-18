package internal

import "net/url"

func ParseURL(value string) (*url.URL, error) {
	parsed, err := url.Parse(value)
	if err != nil {
		return nil, ErrOf(err, "can not parse url %s", value)
	}
	return parsed, nil
}
