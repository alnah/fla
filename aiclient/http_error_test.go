package aiclient

import (
	"errors"
	"testing"
)

func TestHTTPClientError_ErrorString(t *testing.T) {
	testCases := []struct {
		name  string
		hcErr httpClientError
		want  string
	}{
		{
			name: "HappyPath",
			hcErr: httpClientError{
				Provider:  provider("api service"),
				Operation: operation("fetch data"),
				Message:   "failed to build http request",
				Wrapped:   errors.New("malformed url"),
			},
			want: "http client error: operation fetch data failed for api service: failed to build http request: malformed url",
		},
		{
			name:  "OnlyOperation",
			hcErr: httpClientError{Operation: operation("fetch data")},
			want:  "http client error: operation fetch data failed",
		},
		{
			name:  "OnlyProvider",
			hcErr: httpClientError{Provider: provider("api service")},
			want:  "http client error for api service",
		},
		{
			name:  "OnlyMessage",
			hcErr: httpClientError{Message: "bad request"},
			want:  "http client error: bad request",
		},
		{
			name:  "OnlyWrapped",
			hcErr: httpClientError{Wrapped: errors.New("timeout")},
			want:  "http client error: timeout",
		},
		{
			name: "OperationAndProvider",
			hcErr: httpClientError{
				Operation: operation("fetch data"),
				Provider:  provider("api service"),
			},
			want: "http client error: operation fetch data failed for api service",
		},
		{
			name: "OperationAndMessage",
			hcErr: httpClientError{
				Operation: operation("fetch data"),
				Message:   "something went wrong",
			},
			want: "http client error: operation fetch data failed: something went wrong",
		},
		{
			name: "ProviderAndMessage",
			hcErr: httpClientError{
				Provider: provider("api service"),
				Message:  "something went wrong",
			},
			want: "http client error for api service: something went wrong",
		},
		{
			name: "OperationAndWrapped",
			hcErr: httpClientError{
				Operation: operation("fetch data"),
				Wrapped:   errors.New("malformed url"),
			},
			want: "http client error: operation fetch data failed: malformed url",
		},
		{
			name: "ProviderAndWrapped",
			hcErr: httpClientError{
				Provider: provider("api service"),
				Wrapped:  errors.New("malformed url"),
			},
			want: "http client error for api service: malformed url",
		},
		{
			name: "MessageAndWrapped",
			hcErr: httpClientError{
				Message: "something went wrong",
				Wrapped: errors.New("malformed url"),
			},
			want: "http client error: something went wrong: malformed url",
		},
		{
			name:  "Nothing",
			hcErr: httpClientError{},
			want:  "http client error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.hcErr.Error()
			if got != tc.want {
				t.Errorf("want %q, got %q", tc.want, got)
			}
		})
	}
}

func TestHTTPClientError_Unwrap(t *testing.T) {
	cause := errors.New("timeout")
	err := &httpClientError{Wrapped: cause}

	if unwrapped := err.Unwrap(); unwrapped != cause {
		t.Errorf("want %v, got %v", cause, unwrapped)

	}
}
