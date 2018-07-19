package md2http

import (
	"bytes"
	"testing"

	"github.com/pkg/errors"
)

func TestCovert(t *testing.T) {
	var testsCases = []struct {
		name         string
		markdown     string
		expectedErr  error
		expectedHTML string
	}{
		{
			name: "simple",
			markdown: `
## hello, world!

this is a simple example
`,
			expectedErr: nil,
			expectedHTML: `<h2>hello, world!</h2>

<p>this is a simple example</p>
`,
		},
	}

	for _, tc := range testsCases {
		t.Run(tc.name, func(t *testing.T) {
			r := bytes.NewBuffer([]byte(tc.markdown))
			var w bytes.Buffer
			err := Convert(r, &w)

			if tc.expectedErr == nil && err != nil {
				t.Errorf("expected no error, got= %v", err)
			}
			if tc.expectedErr != nil && err == nil {
				t.Errorf("expected error= %v, got none", tc.expectedErr)
			}
			if want, got := tc.expectedErr, err; want != nil && got != nil && errors.Cause(got).Error() != want.Error() {
				t.Errorf("expected error= %v, got= %v", want, got)
			}
			if want, got := tc.expectedHTML, w.String(); want != got {
				t.Errorf("wrong output\n  want= %q\n   got= %q", want, got)
			}
		})
	}
}
