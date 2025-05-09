package prompt

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"strings"
	"text/template"
)

//go:embed */*/*.md
var promptFS embed.FS

// level represents a CEFR level.
type level string

func (l level) String() string { return string(l) }

func (l level) IsValid(path string) error {
	switch l {
	case A1, A2, B1, B2, C1, C2:
		return nil
	default:
		return fmt.Errorf("invalid level %q in %s", l, path)

	}
}

const (
	A1 level = "a1"
	A2 level = "a2"
	B1 level = "b1"
	B2 level = "b2"
	C1 level = "c1"
	C2 level = "c2"
)

// language represents a supported language code.
type language string

func (l language) String() string { return string(l) }

func (l language) IsValid(path string) error {
	switch l {
	case FR, PT:
		return nil
	default:
		return fmt.Errorf("invalid lang %q in %s", l, path)
	}
}

const (
	FR language = "fr"
	PT language = "pt"
)

type promptID string

func (p promptID) String() string { return string(p) }

const (
	SynthSpeech promptID = "synthspeech"
)

type meta struct {
	ID      promptID `json:"id"`
	Lang    language `json:"lang"`
	Level   level    `json:"level"`
	Version int      `json:"version"`
	Tags    []string `json:"tags"`
	Vars    []string `json:"vars"`
}

type prompt struct {
	Meta     meta
	Template *template.Template
}

type runtime struct {
	prompts map[string]*prompt
}

func Runtime() (*runtime, error) {
	mp := make(map[string]*prompt)

	// (O)1 lookup at runtime
	err := fs.WalkDir(promptFS, ".", func(path string, d fs.DirEntry, _ error) error {
		if d.IsDir() {
			return nil
		}
		raw, err := promptFS.ReadFile(path)
		if err != nil {
			return err
		}

		// split ---json … --- front-matter
		parts := bytes.SplitN(raw, []byte("---"), 3)
		if len(parts) < 3 {
			return fmt.Errorf("invalid front-matter in %s", path)
		}
		var m meta
		header := bytes.TrimPrefix(parts[1], []byte("json"))
		if err := json.Unmarshal(header, &m); err != nil {
			return fmt.Errorf("bad json header in %s: %w", path, err)
		}

		// validate level, and language
		if err := m.Level.IsValid(path); err != nil {
			return err
		}
		if err := m.Lang.IsValid(path); err != nil {
			return err
		}

		// compile template with strict missingkey
		tpl, err := template.New(path).Option("missingkey=error").Parse(string(parts[2]))
		if err != nil {
			return fmt.Errorf("template parse error in %s: %w", path, err)
		}

		// normalize casing
		key := strings.Join([]string{m.Lang.String(), m.Level.String(), m.ID.String()}, "*")
		mp[key] = &prompt{Meta: m, Template: tpl}

		// if we already saw a version for this key, only replace when new > old
		// (O)n of the number of files
		if existing, ok := mp[key]; ok {
			if m.Version <= existing.Meta.Version {
				return nil // skip older or equal version
			}
		}
		// otherwise insert or replace
		mp[key] = &prompt{Meta: m, Template: tpl}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &runtime{prompts: mp}, nil
}

func (r *runtime) Prompt(lg language, lvl level, id promptID, data map[string]any) (string, error) {
	strLg := lg.String()
	strLvl := lvl.String()
	strID := id.String()

	trials := []string{
		strings.Join([]string{strLg, strLvl, strID}, "*"),
		strings.Join([]string{strLg, "", strID}, "*"),
		strings.Join([]string{"", strLvl, strID}, "*"),
		strings.Join([]string{"", "", strID}, "*"),
	}

	for _, key := range trials {
		if p, ok := r.prompts[key]; ok {
			var buf bytes.Buffer
			if err := p.Template.Execute(&buf, data); err != nil {
				return "", err
			}
			return strings.TrimSpace(buf.String()), nil
		}
	}

	return "", fmt.Errorf("prompt %q [lang=%s level=%s] not found", id, lg, lvl)
}
