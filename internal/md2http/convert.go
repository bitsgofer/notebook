package md2http

import (
	"io"
	"io/ioutil"

	"github.com/pkg/errors"
	blackfriday "gopkg.in/russross/blackfriday.v2"
)

func Convert(r io.Reader, w io.Writer) error {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return errors.Wrapf(err, "cannot read data")
	}

	converted := blackfriday.Run(b)
	n, err := w.Write(converted)
	if err != nil {
		return errors.Wrapf(err, "cannot write data")
	}
	if want, got := n, len(converted); want != got {
		return errors.Errorf("only wrote %d/%d bytes", want, got)
	}

	return nil
}
