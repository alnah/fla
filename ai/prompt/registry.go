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

// CEFR levels (plus "_" = any‐level)
type level string

const (
	A1       level = "a1"
	A2       level = "a2"
	B1       level = "b1"
	B2       level = "b2"
	C1       level = "c1"
	C2       level = "c2"
	AnyLevel level = "_" // wildcard‐level
)

// ordered list for falling back to the next-lower
var levelsOrdered = []level{A1, A2, B1, B2, C1, C2}
var levelIndex = map[level]int{A1: 0, A2: 1, B1: 2, B2: 3, C1: 4, C2: 5}

func (l level) String() string { return string(l) }
func (l level) IsValid(path string) error {
	switch l {
	case A1, A2, B1, B2, C1, C2, AnyLevel:
		return nil
	default:
		return fmt.Errorf("invalid level %q in %s", l, path)
	}
}

// Supported languages
type language string

const (
	FR language = "fr"
	PT language = "pt"
)

func (l language) String() string { return string(l) }
func (l language) IsValid(path string) error {
	switch l {
	case FR, PT:
		return nil
	default:
		return fmt.Errorf("invalid lang %q in %s", l, path)
	}
}

// Prompt identifier
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

// Registry builds the registry from the embedded .md files.
func Registry() (*runtime, error) {
	return registry(promptFS)
}

// registry reads every file, keeps only the highest-version per "lang*level*id".
func registry(fsys fs.FS) (*runtime, error) {
	mp := make(map[string]*prompt)

	err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, _ error) error {
		if d.IsDir() {
			return nil
		}
		raw, err := fs.ReadFile(fsys, path)
		if err != nil {
			return err
		}
		parts := bytes.SplitN(raw, []byte("---"), 3)
		if len(parts) < 3 {
			return fmt.Errorf("prompt: invalid front-matter in %s", path)
		}

		var m meta
		header := bytes.TrimPrefix(parts[1], []byte("json"))
		if err := json.Unmarshal(header, &m); err != nil {
			return fmt.Errorf("prompt: bad json header in %s: %w", path, err)
		}
		if err := m.Level.IsValid(path); err != nil {
			return err
		}
		if err := m.Lang.IsValid(path); err != nil {
			return err
		}

		tpl, err := template.New(path).Option("missingkey=error").Parse(string(parts[2]))
		if err != nil {
			return fmt.Errorf("prompt template parse error in %s: %w", path, err)
		}

		key := strings.Join([]string{m.Lang.String(), m.Level.String(), m.ID.String()}, "*")
		if existing, ok := mp[key]; ok && existing.Meta.Version >= m.Version {
			return nil
		}
		mp[key] = &prompt{Meta: m, Template: tpl}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &runtime{prompts: mp}, nil
}

// Prompt looks for:
//  1. exact lang+level+id
//  2. same-lang fallback on next-lower CEFR level
//  3. same-lang wildcard-level ("_")
//  4. error otherwise
func (r *runtime) Prompt(lg language, lvl level, id promptID, data map[string]any) (string, error) {
	lang := lg.String()
	idStr := id.String()
	levelStr := lvl.String()

	// exact match
	exactKey := strings.Join([]string{lang, levelStr, idStr}, "*")
	if p, ok := r.prompts[exactKey]; ok {
		var buf bytes.Buffer
		if err := p.Template.Execute(&buf, data); err != nil {
			return "", err
		}
		return strings.TrimSpace(buf.String()), nil
	}

	// fallback same-language on next-lower level, or wildcard "_" for any
	if idx, found := levelIndex[lvl]; found {
		for i := idx - 1; i >= 0; i-- {
			fbLevel := levelsOrdered[i].String()
			fbKey := strings.Join([]string{lang, fbLevel, idStr}, "*")
			if p, ok := r.prompts[fbKey]; ok {
				var buf bytes.Buffer
				if err := p.Template.Execute(&buf, data); err != nil {
					return "", err
				}
				return strings.TrimSpace(buf.String()), nil
			}
		}
	}

	// not found
	return "", fmt.Errorf("prompt not found: id=%s, lang=%s, level=%s", idStr, lang, levelStr)
}
