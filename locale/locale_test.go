package locale

import (
	"fmt"
	"testing"

	"golang.org/x/text/language"
)

func TestISO6391(t *testing.T) {
	type testCase struct {
		code      ISO6391
		wantValid bool
		wantErr   bool
	}
	cases := []testCase{
		// valid
		{code: "en", wantValid: true, wantErr: false},
		{code: "EN", wantValid: true, wantErr: false},
		{code: "Fr", wantValid: true, wantErr: false},

		// invalid
		{code: "zz", wantValid: false, wantErr: true},
		{code: "e", wantValid: false, wantErr: true},
		{code: "eng", wantValid: false, wantErr: true},
		{code: "1n", wantValid: false, wantErr: true},
		{code: "", wantValid: false, wantErr: true},
	}

	for _, tc := range cases {
		t.Run(string(tc.code), func(t *testing.T) {
			gotStr := tc.code.String()
			if gotStr != string(tc.code) {
				t.Errorf("string: want %q, got %q", tc.code, gotStr)
			}

			gotValid := tc.code.IsValid()
			if gotValid != tc.wantValid {
				t.Errorf("is valid: want %v, got %v", tc.wantValid, gotValid)
			}

			err := tc.code.Validate()
			gotErr := err != nil
			if gotErr != tc.wantErr {
				t.Errorf("validate: want %v, got %v", tc.wantErr, gotErr)
			}
		})
	}
}

func TestIETF(t *testing.T) {
	type testCase struct {
		tag       IETF
		wantValid bool
		wantErr   bool
	}
	cases := []testCase{
		// valid
		{tag: "en-US", wantValid: true, wantErr: false},
		{tag: "fr-FR", wantValid: true, wantErr: false},
		{tag: "pt-BR", wantValid: true, wantErr: false},
		{tag: "en_US", wantValid: true, wantErr: false},
		{tag: "EN-us", wantValid: true, wantErr: false},

		// invalid
		{tag: "xyz", wantValid: false, wantErr: true},
		{tag: "", wantValid: false, wantErr: true},
	}

	for _, tc := range cases {
		t.Run(string(tc.tag), func(t *testing.T) {
			gotStr := tc.tag.String()
			if gotStr != string(tc.tag) {
				t.Errorf("string: want %q, got %q", tc.tag, gotStr)
			}

			gotValid := tc.tag.IsValid()
			if gotValid != tc.wantValid {
				t.Errorf("is valid: want %v, got %v", tc.wantValid, gotValid)
			}

			err := tc.tag.Validate()
			gotErr := err != nil
			if gotErr != tc.wantErr {
				t.Errorf("validate: want %v, got %v", tc.wantErr, gotErr)
			}
		})
	}
}

func TestLang(t *testing.T) {
	type testCase struct {
		lang      Lang
		wantValid bool
		wantErr   bool
	}
	cases := []testCase{
		// valid
		{lang: LangEnUS, wantValid: true, wantErr: false},
		{lang: LangFrFR, wantValid: true, wantErr: false},
		{lang: LangPtBR, wantValid: true, wantErr: false},

		// invalid
		{lang: "en-GB", wantValid: false, wantErr: true},
		{lang: "enus", wantValid: false, wantErr: true},
		{lang: "EN-US", wantValid: false, wantErr: true},
		{lang: "", wantValid: false, wantErr: true},
	}

	for _, tc := range cases {
		t.Run(string(tc.lang), func(t *testing.T) {
			gotStr := tc.lang.String()
			if gotStr != string(tc.lang) {
				t.Errorf("string: want %q, got %q", tc.lang, gotStr)
			}

			gotValid := tc.lang.IsValid()
			if gotValid != tc.wantValid {
				t.Errorf("is valid: want %v, got %v", tc.wantValid, gotValid)
			}

			err := tc.lang.Validate()
			gotErr := err != nil
			if gotErr != tc.wantErr {
				t.Errorf("validate: want %v, got %v", tc.wantErr, gotErr)
			}
		})
	}
}

func TestConversions(t *testing.T) {
	for iso, wantLang := range isoToLang {
		t.Run(fmt.Sprintf("fromISO_%s", iso), func(t *testing.T) {
			got, err := FromISO6391(iso)
			gotErr := err != nil
			if gotErr {
				t.Errorf("fromISO error: want %v, got %v", false, gotErr)
			}
			if got != wantLang {
				t.Errorf("fromISO: want %v, got %v", wantLang, got)
			}
		})
	}

	t.Run("fromISO_unmapped_valid", func(t *testing.T) {
		_, err := FromISO6391("de")
		gotErr := err != nil
		if gotErr != true {
			t.Errorf("fromISO: want %v, got %v", true, gotErr)
		}
	})
	t.Run("fromISO_invalid", func(t *testing.T) {
		_, err := FromISO6391("1d")
		gotErr := err != nil
		if gotErr != true {
			t.Errorf("fromISO: want %v, got %v", true, gotErr)
		}
	})

	for ietf, wantLang := range ietfToLang {
		t.Run(fmt.Sprintf("fromIETF_%s", ietf), func(t *testing.T) {
			got, err := FromIETF(ietf)
			gotErr := err != nil
			if gotErr {
				t.Errorf("fromIETF error: want %v, got %v", false, gotErr)
			}
			if got != wantLang {
				t.Errorf("fromIETF: want %v, got %v", wantLang, got)
			}
		})
	}

	t.Run("fromIETF_unmapped_valid", func(t *testing.T) {
		_, err := FromIETF("es-ES")
		gotErr := err != nil
		if gotErr != true {
			t.Errorf("fromIETF: want %v, got %v", true, gotErr)
		}
	})
	t.Run("fromIETF_invalid", func(t *testing.T) {
		_, err := FromIETF("es_ES")
		gotErr := err != nil
		if gotErr != true {
			t.Errorf("fromIETF: want %v, got %v", true, gotErr)
		}
	})
}

func TestLang_ToISO6391_ToIETF(t *testing.T) {
	for _, l := range []Lang{LangEnUS, LangFrFR, LangPtBR} {
		t.Run(string(l), func(t *testing.T) {
			gotISO := l.ToISO6391()
			wantISO := ISO6391(string(l)[:2])
			if gotISO != wantISO {
				t.Errorf("toISO6391: want %v, got %v", wantISO, gotISO)
			}

			gotIETF := l.ToIETF()
			wantIETF := IETF(l)
			if gotIETF != wantIETF {
				t.Errorf("toIETF: want %v, got %v", wantIETF, gotIETF)
			}
		})
	}
}

func TestDisplayName_Variants(t *testing.T) {
	langs := []Lang{LangEnUS, LangFrFR, LangPtBR}
	for _, l := range langs {
		t.Run(string(l), func(t *testing.T) {
			en := l.DisplayName(LangEnUS)
			fr := l.DisplayName(LangFrFR)
			pt := l.DisplayName(LangPtBR)

			if en == "" {
				t.Errorf("display name en: want non-empty, got empty")
			}
			if fr == "" {
				t.Errorf("display name fr: want non-empty, got empty")
			}
			if pt == "" {
				t.Errorf("display name pt: want non-empty, got empty")
			}
			if l != LangEnUS && en == fr {
				t.Errorf("display name: want English and French to differ, got both %q", en)
			}
		})
	}
}

func TestLanguageName_Shortcuts(t *testing.T) {
	for _, l := range []Lang{LangEnUS, LangFrFR, LangPtBR} {
		t.Run(string(l), func(t *testing.T) {
			eng := l.EnglishName()
			if eng != l.DisplayName(LangEnUS) {
				t.Errorf("english name: want %q, got %q", l.DisplayName(LangEnUS), eng)
			}

			fr := l.FrenchName()
			if fr != l.DisplayName(LangFrFR) {
				t.Errorf("french name: want %q, got %q", l.DisplayName(LangFrFR), fr)
			}

			pt := l.PortugueseName()
			if pt != l.DisplayName(LangPtBR) {
				t.Errorf("portuguese name: want %q, got %q", l.DisplayName(LangPtBR), pt)
			}
		})
	}
}

func TestLanguagePackageIntegrity(t *testing.T) {
	for _, raw := range []string{string(LangEnUS), string(LangFrFR), string(LangPtBR)} {
		t.Run("make_"+raw, func(t *testing.T) {
			tag := language.Make(raw)
			if tag.IsRoot() {
				t.Errorf("make: want non-root, got root")
			}
		})
	}
}

func TestLang_Set_Type(t *testing.T) {
	type testCase struct {
		input     string
		wantErr   bool
		wantValue Lang
	}

	cases := []testCase{
		// valid
		{input: "en-US", wantErr: false, wantValue: LangEnUS},
		{input: "fr-FR", wantErr: false, wantValue: LangFrFR},
		{input: "pt-BR", wantErr: false, wantValue: LangPtBR},

		// invalid
		{input: "en-GB", wantErr: true},
		{input: "enus", wantErr: true},
		{input: "", wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			var l Lang
			err := l.Set(tc.input)
			gotErr := err != nil
			if gotErr != tc.wantErr {
				t.Errorf("set: wantErr %v, got %v", tc.wantErr, gotErr)
			}
			if !tc.wantErr {
				// only check value when no error
				if l != tc.wantValue {
					t.Errorf("set: want %v, got %v", tc.wantValue, l)
				}
			}
		})
	}

	// Test Type() always returns "Lang"
	var l Lang
	gotType := l.Type()
	const wantType = "Lang"
	if gotType != wantType {
		t.Errorf("type: want %q, got %q", wantType, gotType)
	}
}
