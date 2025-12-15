package configuration

import (
	"fmt"
	"maps"
)

// HealthCheckConfig is the configuration for
// individual health checks or a default one.
type HealthCheckConfig struct {
	// Option corresponds to "option httpchk"
	// If Method is empty, defaults to "OPTIONS"
	Option HealthCheckOption `json:"option" yaml:"option" toml:"option"`

	// Send corresponds to "http-check send" directives
	Send []HealthCheckSend `json:"send,omitempty" yaml:"send,omitempty" toml:"send,omitempty"`

	// Expect corresponds to "http-check expect" directives
	Expect []HealthCheckExpect `json:"expect,omitempty" yaml:"expect,omitempty" toml:"expect,omitempty"`
}

// HealthCheckOption represents the "option httpchk" directive
type HealthCheckOption struct {
	// Enabled indicates if http health checking is enabled
	Enabled bool `json:"enabled" yaml:"enabled" toml:"enabled"`

	// Method is the HTTP method (GET, POST, HEAD, OPTIONS, etc.)
	Method string `json:"method,omitempty" yaml:"method,omitempty" toml:"method,omitempty"`

	// URI is the path to check
	URI string `json:"uri,omitempty" yaml:"uri,omitempty" toml:"uri,omitempty"`

	// Version is the HTTP version (e.g., "HTTP/1.0", "HTTP/1.1")
	Version string `json:"version,omitempty" yaml:"version,omitempty" toml:"version,omitempty"`
}

// HealthCheckSend represents an "http-check send" directive
type HealthCheckSend struct {
	// Method is the HTTP method
	Method string `json:"method,omitempty" yaml:"method,omitempty" toml:"method,omitempty"`

	// URI is the request URI
	URI string `json:"uri,omitempty" yaml:"uri,omitempty" toml:"uri,omitempty"`

	// Version is the HTTP version
	Version string `json:"version,omitempty" yaml:"version,omitempty" toml:"version,omitempty"`

	// Headers are additional headers to send
	Headers map[string]string `json:"headers,omitempty" yaml:"headers,omitempty" toml:"headers,omitempty"`

	// Body is the request body (optional)
	Body string `json:"body,omitempty" yaml:"body,omitempty" toml:"body,omitempty"`
}

// HealthCheckExpect represents an "http-check expect" directive
type HealthCheckExpect struct {
	// Type is the expectation type: "status", "string", "rstring", "rstatus"
	Type string `json:"type" yaml:"type" toml:"type"`

	// Match determines if we're matching or negating (!match)
	Match bool `json:"match" yaml:"match" toml:"match"`

	// Value is the value to match against
	// For status: e.g., "200", "200-299"
	// For string/rstring: the string or regex pattern
	Value string `json:"value" yaml:"value" toml:"value"`
}

// Copy creates a deep copy of the HealthCheckConfig
func (h *HealthCheckConfig) Copy() *HealthCheckConfig {
	if h == nil {
		return nil
	}

	copyConfig := &HealthCheckConfig{
		Option: HealthCheckOption{
			Enabled: h.Option.Enabled,
			Method:  h.Option.Method,
			URI:     h.Option.URI,
			Version: h.Option.Version,
		},
	}

	if len(h.Send) > 0 {
		copyConfig.Send = make([]HealthCheckSend, len(h.Send))
		for i, send := range h.Send {
			copySend := HealthCheckSend{
				Method:  send.Method,
				URI:     send.URI,
				Version: send.Version,
				Body:    send.Body,
			}
			if len(send.Headers) > 0 {
				copySend.Headers = make(map[string]string)
				maps.Copy(copySend.Headers, send.Headers)
			}
			copyConfig.Send[i] = copySend
		}
	}

	if len(h.Expect) > 0 {
		copyConfig.Expect = make([]HealthCheckExpect, len(h.Expect))
		for i, expect := range h.Expect {
			copyConfig.Expect[i] = HealthCheckExpect{
				Type:  expect.Type,
				Match: expect.Match,
				Value: expect.Value,
			}
		}
	}

	return copyConfig
}

// Example usage and string generation
func (h *HealthCheckConfig) String() string {
	if !h.Option.Enabled {
		return ""
	}

	var result string

	// Generate option httpchk
	result += "  option httpchk"
	if h.Option.Method != "" {
		result += fmt.Sprintf(" %s", h.Option.Method)
	}
	if h.Option.URI != "" {
		result += fmt.Sprintf(" %s", h.Option.URI)
	}
	if h.Option.Version != "" {
		result += fmt.Sprintf(" %s", h.Option.Version)
	}
	result += "\n"

	// Generate http-check send directives
	for _, send := range h.Send {
		result += "  http-check send"
		if send.Method != "" {
			result += fmt.Sprintf(" meth %s", send.Method)
		}
		if send.URI != "" {
			result += fmt.Sprintf(" uri %s", send.URI)
		}
		if send.Version != "" {
			result += fmt.Sprintf(" ver %s", send.Version)
		}
		for key, val := range send.Headers {
			result += fmt.Sprintf(" hdr %s %s", key, val)
		}
		if send.Body != "" {
			result += fmt.Sprintf(" body %s", send.Body)
		}
		result += "\n"
	}

	// Generate http-check expect directives
	for _, expect := range h.Expect {
		result += "  http-check expect"
		if !expect.Match {
			result += " !"
		}
		result += fmt.Sprintf(" %s %s", expect.Type, expect.Value)
		result += "\n"
	}

	return result
}
