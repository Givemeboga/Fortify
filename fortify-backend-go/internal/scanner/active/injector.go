package active

import (
	"net/url"
)

// InjectPayload mirrors Python injector.py: for every query parameter, build a
// copy of the URL with ONLY that parameter replaced by the payload.
// Returns param -> injected URL.
func InjectPayload(rawURL, payload string) map[string]string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return map[string]string{}
	}
	q, err := url.ParseQuery(u.RawQuery)
	if err != nil || len(q) == 0 {
		return map[string]string{}
	}
	out := make(map[string]string, len(q))
	for param := range q {
		clone, _ := url.Parse(rawURL)
		cq, _ := url.ParseQuery(clone.RawQuery)
		cq.Set(param, payload)
		clone.RawQuery = cq.Encode()
		out[param] = clone.String()
	}
	return out
}
