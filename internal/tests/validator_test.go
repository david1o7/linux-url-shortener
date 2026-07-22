package tests

import (
	"Linux-url-shortener/internal/validator"
	"context"
	"errors"
	"net"
	"net/http"
	"testing"
)

func TestValidate(t *testing.T){

	tests:=[]struct{

		name string

		url string

		response int

		err error

		expected bool

	}{
		{
			name:"Valid URL",

			url:"https://google.com",

			response:http.StatusOK,

			expected:true,
		}, 

		{
			name:"404",

			url:"https://babble2234.com",

			response:http.StatusNotFound,

			expected:false,
		},

		{
			name:"Timeout",

			url:"https://TimeOutError.com",

			err:context.DeadlineExceeded,

			expected:false,
		},

		{
			name:"FTP",

			url:"ftp://google.com",

			expected:false,
		},

		{
			name:"Invalid URL",

			url:"hello",

			expected:false,
		},

	}

	for _,tt:=range tests{

		client:=&MockClient{

			Response:&http.Response{

				StatusCode:tt.response,

				Body:http.NoBody,
			},

			Err:tt.err,
		}

		resolver := &MockResolver{
			IPs: []net.IP{
				net.ParseIP("8.8.8.8"),
			},
		}

		v:=validator.NewURLValidator(client, resolver, 10)

		got:=v.Validate(tt.url)

		if got!=tt.expected{

			t.Fatalf("%s expected %v got %v",
				tt.name,
				tt.expected,
				got,
			)

		}

	}
}

func TestRejectLoopback(t *testing.T){

	v:=validator.NewURLValidator(nil, &MockResolver{
		IPs: []net.IP{
			net.ParseIP("127.0.0.1"),
		},
	}, 10)

	if v.Validate("http://127.0.0.1"){

		t.Fatal("expected loopback to fail")

	}

}

func TestRejectPrivateIP(t *testing.T){

	v:= validator.NewURLValidator(nil, &MockResolver{
		IPs: []net.IP{
			net.ParseIP("192.168.1.1"),
		},
	}, 10)

	if v.Validate("http://192.168.1.1"){

		t.Fatal("expected private ip to fail")

	}

}

func TestRejectUnsupportedScheme(t *testing.T){

	v:=validator.NewURLValidator(nil, &MockResolver{
		IPs: []net.IP{
			net.ParseIP("8.8.8.8"),
		},
	}, 10)

	if v.Validate("ftp://example.com"){

		t.Fatal("expected ftp to fail")

	}

}

func TestNetworkFailure(t *testing.T){

	v:=validator.NewURLValidator(nil, &MockResolver{
		Err: errors.New("Dns failed"),
	}, 10)

	if v.Validate("https://google.com"){

		t.Fatal("expected validation failure")

	}

}