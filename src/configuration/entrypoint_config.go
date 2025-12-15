package configuration

import "fmt"

// HeaderConfig represents a config for a specific header.
type HeaderConfig struct {
	// Value represents the value of the header
	Value string `json:"value" yaml:"value" toml:"value"`

	// AppendValue determines if the value should be appended or not.
	AppendValue bool `json:"appendValue" yaml:"append_value" toml:"append_value"`
}

// EntrypointConfig represents extra
// headers to apply on backends
// for specific entrypoints
type EntrypointConfig struct {
	// RequestHeaders is the request headers
	// to add to each backend request.
	RequestHeaders map[string]*HeaderConfig `json:"requestHeaders" yaml:"request_headers" toml:"request_headers"`
}

func (c *EntrypointConfig) String() string {
	var result string

	for key, value := range c.RequestHeaders {
		if value.AppendValue {
			result += fmt.Sprintf("  http-request add-header %s %s", key, value.Value)
		} else {
			result += fmt.Sprintf("  http-request set-header %s %s", key, value.Value)
		}

		result += "\n"
	}

	return result
}
