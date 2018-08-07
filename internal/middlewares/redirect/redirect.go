package redirection

import (
	"github.com/pkg/errors"
)

func Validate(rds Redirections, servableDomains ...string) error {
	servable := make(map[string]struct{}, len(servableDomains))
	for _, d := range servableDomains {
		servable[d] = struct{}{}
	}
	for _, rd := range rds {
		if rd.FromURL.Scheme != "https" {
			return errors.Errorf("cannot redirect from %v, scheme must be HTTPS", rd.FromURL.String())
		}
		if _, exist := servable[rd.FromURL.Host]; !exist {
			return errors.Errorf("cannot redirect from %v, not serving the domain", rd.FromURL.String())
		}
	}

	return nil
}
