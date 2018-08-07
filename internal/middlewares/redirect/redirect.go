package redirect

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/pkg/errors"
)

type Redirection struct {
	FromURL, ToURL url.URL
}

type Redirections []Redirection

func (rds *Redirections) String() string {
	if len(*rds) < 1 {
		return "{}"
	}

	var buf bytes.Buffer
	buf.WriteString("{")
	for _, rd := range *rds {
		from, to := rd.FromURL, rd.ToURL
		buf.WriteString(fmt.Sprintf("%v => %v, ", from.String(), to.String()))
	}
	buf.WriteString("}")

	return buf.String()
}

func (rds *Redirections) Set(strs string) error {
	if len(strs) < 1 {
		return nil
	}

	pairs := strings.Split(strs, ",")
	res := make([]Redirection, 0, len(pairs))
	for _, str := range pairs {
		parts := strings.Split(str, "=>")
		if len(parts) != 2 {
			return errors.Errorf("%q is not a valid redirection", str)
		}

		from, err := url.Parse(parts[0])
		if err != nil {
			return errors.Errorf("URL to redirect from, %q, is not valid", parts[0])
		}
		to, err := url.Parse(parts[1])
		if err != nil {
			return errors.Errorf("URL to redirect to, %q, is not valid", parts[1])
		}

		res = append(res, Redirection{FromURL: *from, ToURL: *to})
	}

	*rds = res
	return nil
}

func NewHandler(next http.Handler, rds Redirections, domains ...string) (http.Handler, error) {
	servable := make(map[string]struct{}, len(domains))
	for _, d := range domains {
		servable[d] = struct{}{}
	}

	index := make(map[string]string, len(rds))
	for _, rd := range rds {
		if !(rd.FromURL.Scheme == "" || rd.FromURL.Scheme == "http" || rd.FromURL.Scheme == "https") {
			return nil, errors.Errorf("cannot redirect from URL with %s scheme", rd.FromURL.Scheme)
		}
		if _, exist := servable[rd.FromURL.Host]; !exist {
			return nil, errors.Errorf("cannot redirect from %v, not serving the domain", rd.FromURL.String())
		}

		k := redirectKey(rd.FromURL.Host, rd.FromURL.Path, rd.FromURL.RawQuery)
		index[k] = rd.ToURL.String()
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		host, path := r.URL.Host, r.URL.Path
		if len(r.URL.Host) < 1 {
			host = r.Host
		}

		k := redirectKey(host, path, r.URL.RawQuery)
		rdURL, exist := index[k]
		if !exist {
			next.ServeHTTP(w, r)
			return
		}

		http.Redirect(w, r, rdURL, http.StatusMovedPermanently)
	}

	return http.HandlerFunc(handler), nil
}

func redirectKey(host, path, rawQuery string) string {
	return fmt.Sprintf("%s%s?%s", host, path, rawQuery)
}
