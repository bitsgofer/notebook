package redirection

import (
	"flag"
	"net/url"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
)

var _ flag.Value = (*Redirections)(nil)

var (
	url1, _   = url.Parse("http://subdomain.example.com")
	url2, _   = url.Parse("https://example.com/path?with=query")
	url3, _   = url.Parse("host.only")
	url4, _   = url.Parse("another.host:80")
	url5, _   = url.Parse("https://valid.domain/path?q=v")
	ftpURL, _ = url.Parse("ftp://file/path")
)

func TestRedirectionsString(t *testing.T) {
	var testCases = []struct {
		val      Redirections
		expected string
	}{
		{
			Redirections{},
			"{}",
		},
		{
			Redirections{
				Redirection{FromURL: *url1, ToURL: *url2},
				Redirection{FromURL: *url3, ToURL: *url4},
			},
			"{http://subdomain.example.com => https://example.com/path?with=query, host.only => another.host:80, }",
		},
	}

	for _, tc := range testCases {
		actual := tc.val.String()
		if want, got := tc.expected, actual; want != got {
			t.Errorf("wrong result;\n  want= %q\n   got= %q", want, got)
		}
	}
}

func TestRedirectionsSet(t *testing.T) {
	var testCases = []struct {
		name        string
		str         string
		expected    Redirections
		expectedErr error
	}{
		{
			"single redirection",
			"host.only=>https://example.com/path?with=query",
			Redirections{
				Redirection{FromURL: *url3, ToURL: *url2},
			},
			nil,
		},
		{
			"multiple redirection",
			"host.only=>https://example.com/path?with=query,another.host:80=>http://subdomain.example.com",
			Redirections{
				Redirection{FromURL: *url3, ToURL: *url2},
				Redirection{FromURL: *url4, ToURL: *url1},
			},
			nil,
		},
		{
			"no redirection",
			"",
			nil,
			nil,
		},
		{
			"bad redirection",
			"host.only== https://example.com/path?with=query",
			nil,
			errors.New("\"host.only== https://example.com/path?with=query\" is not a valid redirection"),
		},
		{
			"bad fromURL",
			"host.only:// =>https://example.com/path?with=query",
			nil,
			errors.New("URL to redirect from, \"host.only:// \", is not valid"),
		},
		{
			"bad toURL",
			"host.only=> https://example.com/path?with=query",
			nil,
			errors.New("URL to redirect to, \" https://example.com/path?with=query\", is not valid"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var rds Redirections
			err := rds.Set(tc.str)
			if tc.expectedErr == nil && err != nil {
				t.Errorf("expected no error, got= %v", err)
			}
			if tc.expectedErr != nil && err == nil {
				t.Errorf("want err= %v, got none", tc.expectedErr)
			}
			if tc.expectedErr != nil && err != nil && errors.Cause(err).Error() != tc.expectedErr.Error() {
				t.Errorf("wrong error\n  want= %v\n   got %v", tc.expectedErr, err)
			}
			if want, got := tc.expected, rds; !cmp.Equal(want, got) {
				t.Errorf("wrong flag set\n  want= %#v\n   got= %#v\n  diff= %v", want, got, cmp.Diff(want, got))
			}
		})
	}
}

func TestValidate(t *testing.T) {
	var testCases = []struct {
		name            string
		redirections    Redirections
		servableDomains []string
		expectedErr     error
	}{
		{
			"multiple redirections",
			Redirections{
				Redirection{FromURL: *url1, ToURL: *url5},
				Redirection{FromURL: *url2, ToURL: *url5},
			},
			[]string{"example.com", "subdomain.example.com"},
			nil,
		},
		{
			"no redirections",
			Redirections{},
			[]string{"example.com", "subdomain.example.com"},
			nil,
		},
		{
			"not serving domain",
			Redirections{
				Redirection{FromURL: *url1, ToURL: *url5},
			},
			[]string{},
			errors.New("cannot redirect from http://subdomain.example.com, not serving the domain"),
		},
		{
			"not HTTP or HTTPS",
			Redirections{
				Redirection{FromURL: *ftpURL, ToURL: *url5},
			},
			[]string{"example.com", "subdomain.example.com"},
			errors.New("cannot redirect from ftp://file/path, scheme must be HTTP or HTTPS"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := Validate(tc.redirections, tc.servableDomains...)
			if tc.expectedErr == nil && err != nil {
				t.Errorf("want no error, got= %v", err)
			}
			if tc.expectedErr != nil && err == nil {
				t.Errorf("want err= %v, got none", tc.expectedErr)
			}
			if tc.expectedErr != nil && err != nil && errors.Cause(err).Error() != tc.expectedErr.Error() {
				t.Errorf("wrong error\n  want= %v\n   got= %v", tc.expectedErr, err)
			}
		})
	}
}
