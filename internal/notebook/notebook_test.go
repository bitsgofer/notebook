package notebook

import (
	"os"
	"testing"

	"github.com/pkg/errors"
)

func TestNewArticle(t *testing.T) {
	var testCases = []struct {
		name        string
		path        string
		expectedErr error
	}{
		{
			name:        "normal",
			path:        "testdata/nb1/test.md",
			expectedErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			info, err := os.Stat(tc.path) // follow symlink
			if err != nil {
				t.Fatalf("cannot get stat of %q", tc.path)
			}

			_, err = newArticle(tc.path, info)
			if tc.expectedErr == nil && err != nil {
				t.Errorf("expected no error, got= %v", err)
			}
			if tc.expectedErr != nil && err == nil {
				t.Errorf("expected error= %v, got none", tc.expectedErr)
			}
			if want, got := tc.expectedErr, err; want != nil && got != nil && errors.Cause(got).Error() != want.Error() {
				t.Errorf("expected error= %v, got= %v", want, got)
			}
		})
	}
}

func TestNewNotebook(t *testing.T) {
	var testCases = []struct {
		name        string
		path        string
		expectedErr error
	}{
		{
			name:        "normal",
			path:        "testdata/nb1",
			expectedErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewNotebook(tc.path)
			if tc.expectedErr == nil && err != nil {
				t.Errorf("expected no error, got= %v", err)
			}
			if tc.expectedErr != nil && err == nil {
				t.Errorf("expected error= %v, got none", tc.expectedErr)
			}
			if want, got := tc.expectedErr, err; want != nil && got != nil && errors.Cause(got).Error() != want.Error() {
				t.Errorf("expected error= %v, got= %v", want, got)
			}
		})
	}
}
