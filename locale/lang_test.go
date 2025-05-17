package locale

import (
	"testing"
)

/********* Unit Test *********/

func TestLang_Stringer(t *testing.T) {
	testCases := []struct {
		name string
		lang Lang
		want string
	}{
		{name: "FR", lang: FR, want: "fr"},
		{name: "PT", lang: PT, want: "pt"},
		{name: "EN", lang: EN, want: "en"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.lang.String() != tc.want {
				t.Errorf("string: want %q, got %q", tc.want, tc.lang.String())
			}
		})
	}
}

func TestLang_Validation(t *testing.T) {
	testCases := []struct {
		name    string
		lang    Lang
		wantErr bool
	}{
		{name: "ValidateFR", lang: FR, wantErr: false},
		{name: "ValidatePT", lang: PT, wantErr: false},
		{name: "ValidateEN", lang: EN, wantErr: false},
		{name: "InvalidateZZ", lang: Lang("zz"), wantErr: true},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.lang.Validate()
			if tc.wantErr {
				if err == nil {
					t.Errorf("validation: want err, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("validation: want no err, got %v", err)
				}
			}
		})
	}
}
