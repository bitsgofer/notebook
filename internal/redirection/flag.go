package redirection

import (
	"bytes"
	"fmt"
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
