package aiclient

import (
	"errors"
	"testing"
)

func TestHTTPClientError_ErrorString(t *testing.T) {
	testCases := []struct {
		name  string
		hcErr HTTPClientError
		want  string
	}{
		{
			name: "HappyPath",
			hcErr: HTTPClientError{
				Provider:  Provider("api service"),
				Operation: Operation("fetch data"),
				Message:   "failed to build http request",
				Wrapped:   errors.New("malformed url"),
			},
			want: "http client error: operation fetch data failed for api service: failed to build http request: malformed url",
		},
		{
			name:  "OnlyOperation",
			hcErr: HTTPClientError{Operation: Operation("fetch data")},
			want:  "http client error: operation fetch data failed",
		},
		{
			name:  "OnlyProvider",
			hcErr: HTTPClientError{Provider: Provider("api service")},
			want:  "http client error for api service",
		},
		{
			name:  "OnlyMessage",
			hcErr: HTTPClientError{Message: "bad request"},
			want:  "http client error: bad request",
		},
		{
			name:  "OnlyWrapped",
			hcErr: HTTPClientError{Wrapped: errors.New("timeout")},
			want:  "http client error: timeout",
		},
		{
			name: "OperationAndProvider",
			hcErr: HTTPClientError{
				Operation: Operation("fetch data"),
				Provider:  Provider("api service"),
			},
			want: "http client error: operation fetch data failed for api service",
		},
		{
			name: "OperationAndMessage",
			hcErr: HTTPClientError{
				Operation: Operation("fetch data"),
				Message:   "something went wrong",
			},
			want: "http client error: operation fetch data failed: something went wrong",
		},
		{
			name: "ProviderAndMessage",
			hcErr: HTTPClientError{
				Provider: Provider("api service"),
				Message:  "something went wrong",
			},
			want: "http client error for api service: something went wrong",
		},
		{
			name: "OperationAndWrapped",
			hcErr: HTTPClientError{
				Operation: Operation("fetch data"),
				Wrapped:   errors.New("malformed url"),
			},
			want: "http client error: operation fetch data failed: malformed url",
		},
		{
			name: "ProviderAndWrapped",
			hcErr: HTTPClientError{
				Provider: Provider("api service"),
				Wrapped:  errors.New("malformed url"),
			},
			want: "http client error for api service: malformed url",
		},
		{
			name: "MessageAndWrapped",
			hcErr: HTTPClientError{
				Message: "something went wrong",
				Wrapped: errors.New("malformed url"),
			},
			want: "http client error: something went wrong: malformed url",
		},
		{
			name:  "Nothing",
			hcErr: HTTPClientError{},
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
	err := &HTTPClientError{Wrapped: cause}

	if unwrapped := err.Unwrap(); unwrapped != cause {
		t.Errorf("want %v, got %v", cause, unwrapped)

	}
}
