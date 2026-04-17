package render

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ExquisiteCore/cnki-search/internal/model"
)

// Search dispatches a SearchResult to the chosen format.
func Search(r *model.SearchResult, format string) (string, error) {
	switch strings.ToLower(format) {
	case "", "json":
		return marshalJSON(r)
	case "table":
		return searchAsTable(r), nil
	case "citation":
		return searchAsCitation(r), nil
	case "markdown", "md":
		return searchAsMarkdown(r), nil
	default:
		return "", fmt.Errorf("unknown format: %s", format)
	}
}

// Detail dispatches a Detail to the chosen format.
func Detail(d *model.Detail, format string) (string, error) {
	switch strings.ToLower(format) {
	case "", "json":
		return marshalJSON(d)
	case "markdown", "md":
		return detailAsMarkdown(d), nil
	case "citation":
		return detailAsCitation(d), nil
	case "table":
		// A single-paper table is just markdown's info list.
		return detailAsMarkdown(d), nil
	default:
		return "", fmt.Errorf("unknown format: %s", format)
	}
}

// References dispatches a refs slice to the chosen format.
func References(refs []model.Reference, format string) (string, error) {
	switch strings.ToLower(format) {
	case "", "json":
		return marshalJSON(refs)
	case "citation", "table", "markdown", "md":
		var sb strings.Builder
		for _, r := range refs {
			fmt.Fprintf(&sb, "[%d] %s\n", r.Seq, r.Text)
		}
		return strings.TrimRight(sb.String(), "\n"), nil
	default:
		return "", fmt.Errorf("unknown format: %s", format)
	}
}

func marshalJSON(v any) (string, error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}
