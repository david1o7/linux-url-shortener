package tests

import (
	"Linux-url-shortener/internal/tests/mocks"
	"Linux-url-shortener/internal/validator"
	"context"
	"errors"
	"net"
	"net/http"
	"testing"
)

func TestValidate(t *testing.T) {

	tests := []struct {
		name string

		url string

		response int

		err error

		expected bool
	}{
		{
			name: "Valid URL",

			url: "https://google.com",

			response: http.StatusOK,

			expected: true,
		},

		{
			name: "404",

			url: "https://babble2234.com",

			response: http.StatusNotFound,

			expected: false,
		},

		{
			name: "Timeout",

			url: "https://TimeOutError.com",

			err: context.DeadlineExceeded,

			expected: false,
		},

		{
			name: "FTP",

			url: "ftp://google.com",

			expected: false,
		},

		{
			name: "Invalid URL",

			url: "hello",

			expected: false,
		},
	}

	for _, tt := range tests {

		client := &mocks.MockClient{

			Response: &http.Response{

				StatusCode: tt.response,

				Body: http.NoBody,
			},

			Err: tt.err,
		}

		resolver := &mocks.MockResolver{
			IPs: []net.IP{
				net.ParseIP("8.8.8.8"),
			},
		}

		v := validator.NewURLValidator(client, resolver, 10)

		got := v.Validate(tt.url)

		if got != tt.expected {

			t.Fatalf("%s expected %v got %v",
				tt.name,
				tt.expected,
				got,
			)

		}

	}
}

func TestRejectLoopback(t *testing.T) {

	v := validator.NewURLValidator(nil, &mocks.MockResolver{
		IPs: []net.IP{
			net.ParseIP("127.0.0.1"),
		},
	}, 10)

	if v.Validate("http://127.0.0.1") {

		t.Fatal("expected loopback to fail")

	}

}

func TestRejectPrivateIP(t *testing.T) {

	v := validator.NewURLValidator(nil, &mocks.MockResolver{
		IPs: []net.IP{
			net.ParseIP("192.168.1.1"),
		},
	}, 10)

	if v.Validate("http://192.168.1.1") {

		t.Fatal("expected private ip to fail")

	}

}

func TestRejectUnsupportedScheme(t *testing.T) {

	v := validator.NewURLValidator(nil, &mocks.MockResolver{
		IPs: []net.IP{
			net.ParseIP("8.8.8.8"),
		},
	}, 10)

	if v.Validate("ftp://example.com") {

		t.Fatal("expected ftp to fail")

	}

}

func TestNetworkFailure(t *testing.T) {

	v := validator.NewURLValidator(nil, &mocks.MockResolver{
		Err: errors.New("Dns failed"),
	}, 10)

	if v.Validate("https://google.com") {

		t.Fatal("expected validation failure")

	}

}

func TestRejectMulticastIP(t *testing.T) {
	v := validator.NewURLValidator(
		nil,
		&mocks.MockResolver{
			IPs: []net.IP{
				net.ParseIP("224.0.0.1"),
			},
		},
		10,
	)

	if v.Validate("http://example.com") {
		t.Fatal("expected multicast IP to fail")
	}
}

func TestRejectUnspecifiedIP(t *testing.T) {
	v := validator.NewURLValidator(
		nil,
		&mocks.MockResolver{
			IPs: []net.IP{
				net.ParseIP("0.0.0.0"),
			},
		},
		10,
	)

	if v.Validate("http://example.com") {
		t.Fatal("expected unspecified IP to fail")
	}
}

func TestValidateRedirectResponse(t *testing.T) {
	client := &mocks.MockClient{
		Response: &http.Response{
			StatusCode: http.StatusFound,
			Body:       http.NoBody,
		},
	}

	resolver := &mocks.MockResolver{
		IPs: []net.IP{
			net.ParseIP("8.8.8.8"),
		},
	}

	v := validator.NewURLValidator(
		client,
		resolver,
		10,
	)

	if !v.Validate("https://example.com") {
		t.Fatal("expected 3xx response to be accepted")
	}
}

func TestValidateServerError(t *testing.T) {
	client := &mocks.MockClient{
		Response: &http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       http.NoBody,
		},
	}

	resolver := &mocks.MockResolver{
		IPs: []net.IP{
			net.ParseIP("8.8.8.8"),
		},
	}

	v := validator.NewURLValidator(
		client,
		resolver,
		10,
	)

	if v.Validate("https://example.com") {
		t.Fatal("expected 500 response to fail")
	}
}
