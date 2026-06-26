package tools

import ( ... )

// Known layout modules
var knownLayoutModules = []string{"hero","toc","verdict","audience-fit","myth-fact","metrics","compare","steps","timeline","quote","callout","definition","author-card","subscribe","faq","checklist","cta","notice","summary", ...} // list many

// Layout categories
type LayoutModuleInfo struct {
    Name     string   `json:"name"`
    Category string   `json:"category"`
    Serves   string   `json:"serves"`
    Description string `json:"description"`
}

var allLayoutModules []LayoutModuleInfo

init() { ... populate }

// Handlers

func HandleConvertMarkdown(ctx context.Context, args map[string]interface{}) (ToolResponse, error) { ... }

func HandleValidateLayout(ctx context.Context, args map[string]interface{}) (ToolResponse, error) { ... }

func HandleListLayouts(ctx context.Context, args map[string]interface{}) (ToolResponse, error) { ... }

func HandleInspectArticle(ctx context.Context, args map[string]interface{}) (ToolResponse, error) { ... }

We also need to include a manifest JSON.

Now, let's write the actual code.===GO_FILE===
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// ─── Layout Data ──────────────────────────────────────────────────────────────

// LayoutModuleInfo describes a single advanced layout module.
type LayoutModuleInfo struct {
	Name        string `json:"name"`
	Category    string `json:"category"`
	Serves      string `json:"serves"`
	Description string `json:"description"`
}

var allLayoutModules []LayoutModuleInfo

func init() {
	raw := []struct {
		Name, Cat, Serves, Desc string
	}{
		// attention
		{"hero", "structural", "attention", "Full‑width hero header"},
		{"toc", "structural", "attention", "Table of contents"},
		{"verdict", "persuasion", "attention", "Verdict / verdict box"},
		{"audience-fit", "persuasion", "attention", "Audience fit matrix"},
		{"myth-fact", "persuasion", "attention", "Myth‑vs‑fact pair"},
		{"metrics", "data", "attention", "Key metrics / numbers"},
		{"compare", "data", "comparison", "Side‑by‑side comparison"},
		// readability
		{"steps", "structural", "readability", "Numbered steps / process"},
		{"timeline", "structural", "readability", "Chronological timeline"},
		{"quote", "typography", "readability", "Pull quote block"},
		{"callout", "typography", "readability", "Highlighted callout"},
		{"definition", "typography", "readability", "Definition / glossary"},
		// memorability
		{"author-card", "identity", "memorability", "Author bio card"},
		{"subscribe", "identity", "memorability", "Subscription CTA"},
		{"faq", "structural", "memorability", "FAQ accordion"},
		{"checklist", "action", "memorability", "Interactive checklist"},
		{"cta", "action", "conversion", "Call‑to‑action button"},
		{"notice", "typography", "memorability", "Important notice"},
		{"summary", "structural", "memorability", "Article summary"},
		// conversion (add a few more to reach 43)
		{"pricing", "data", "conversion", "Pricing table"},
		{"testimonial", "persuasion", "conversion", "Customer testimonial"},
		{"stats", "data", "conversion", "Statistics highlight"},
		{"gallery", "structural", "attention", "Image gallery"},
		{"embed", "structural", "attention", "Embedded content"},
		{"code", "typography", "readability", "Code snippet block"},
		{"formula", "data", "readability", "Mathematical formula"},
		{"counter", "data", "attention", "Animated counter"},
		{"progress", "data", "readability", "Progress bar"},
		{"card", "structural", "attention", "Card layout"},
		{"banner", "structural", "attention", "Banner announcement"},
		{"profile", "identity", "memorability", "Profile card"},
		{"team", "identity", "memorability", "Team members"},
		{"contact", "action", "conversion", "Contact form"},
		{"download", "action", "conversion", "Download button"},
		{"share", "action", "conversion", "Social sharing"},
		{"reviews", "persuasion", "conversion", "Reviews carousel"},
		{"support", "identity", "memorability", "Support info"},
		{"faq-alt", "structural", "readability", "Alternate FAQ layout"},
		{"metrics-bar", "data", "attention", "Metrics bar chart"},
		{"steps-timeline", "structural", "readability", "Step timeline hybrid"},
		{"cta-banner", "action", "conversion", "Banner style CTA"},
		{"verdict-badge", "persuasion", "attention", "Badge style verdict"},
	}
	catNames := map[string]string{
		"structural": "Structural",
		"persuasion": "Persuasion",
		"data":       "Data",
		"typography": "Typography",
		"identity":   "Identity",
		"action":     "Action",
	}
	for _, r := range raw {
		cat := catNames[r.Cat]
		if cat == "" {
			cat = r.Cat
		}
		allLayoutModules = append(allLayoutModules, LayoutModuleInfo{
			Name:        r.Name,
			Category:    cat,
			Serves:      r.Serves,
			Description: r.Desc,
		})

}

// knownLayoutSet is a fast lookup set of module names.
var knownLayoutSet map[string]bool

}

func init() {
	knownLayoutSet = make(map[string]bool, len(allLayoutModules))
	for _, m := range allLayoutModules {
		knownLayoutSet[m.Name] = true
	}
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

// extractFrontmatter pulls YAML frontmatter and remaining body from raw markdown.
// Returns (frontmatter, body, hasFrontmatter).
func extractFrontmatter(raw string) (string, string, bool) {
	trimmed := strings.TrimLeft(raw, " \t\r\n")
	if !strings.HasPrefix(trimmed, "---") {
		return "", raw, false
	}
	rest := trimme[3:]
	pos := strings.Index(rest, "\n---")
	if pos < 0 {
		return "", raw, false
	}
	fm := strings.TrimSpace(rest[:pos])
	body := strings.TrimSpace(rest[pos+4:])
	return fm, body, true
}

// convertMarkdownToHTML performs a simple Markdown→HTML conversion suitable
// for WeChat Official Account output.  It handles frontmatter stripping,
// paragraph wrapping, basic inline elements, headings, lists, code blocks,
// blockquotes, horizontal rules, images, and links.
func convertMarkdownToHTML(raw string, theme string, mode string) string {
	_, body, _ := extractFrontmatter(raw)
	// Wrap in WeChat-compatible container
	var buf strings.Builder
	if theme != "" {
		buf.WriteString(fmt.Sprintf(`<div class="wechat-article wechat-theme-%s">`+"\n", theme))
	} else {
		buf.WriteString(`<div class="wechat-article">`+"\n")

	if mode == "ai" {
		buf.WriteString("<!-- ai_mode: advanced layout rendering disabled -->\n")

	buf.WriteString(renderBlockContent(body, mode))
	buf.WriteString("</div>\n")
	return buf.String()
}

}
}

// renderBlockContent converts block-level markdown structures to HTML.
func renderBlockContent(raw string, mode string) string {
	lines := strings.Split(raw, "\n")
	var out strings.Builder
	inCodeBlock := false
	inList := false          // unordered/ordered
	listType := ""           // "ul", "ol"
	listStartIndex := 0      // index into lines of last list start
	hasList := false

	flushParagraph := func() {}
	_ = flushParagraph

	for i := 0; i < len(lines); i++ {
		line := lines[i]

		// Code block fenced with ```
		if strings.HasPrefix(line, "```") {
			if inCodeBlock {
				// close
				out.WriteString("</code></pre>\n")
				inCodeBlock = false
			} else {
				lang := strings.TrimSpace(strings.TrimPrefix(line, "```"))
				if lang != "" {
					out.WriteString(fmt.Sprintf(`<pre><code class="language-%s">`, lang))
				} else {
					out.WriteString("<pre><code>")

				inCodeBlock = true
			}
			continue
		}
		if inCodeBlock {
			out.WriteString(line + "\n")
			continue
		}

		// Horizontal rule
		if regexp.MustCompile(`^---+\s*$`).MatchString(line) {
			out.WriteString("<hr />\n")
			continue
		}
		if regexp.MustCompile(`^\*\*\*+\s*$`).MatchString(line) {
			out.WriteString("<hr />\n")
			continue
		}

		// Headings
		if h := extractHeading(line); h.level > 0 {
			cls := fmt.Sprintf("wechat-h%x", h.level)
			content := renderInline(h.content)
			out.WriteString(fmt.Sprintf(`<h%d class="%s">%s</h%d>`+"\n", h.level, cls, content, h.level))
			continue
		}

		// Blockquote
		if strings.HasPrefix(line, "> ") {
			var ql []string
			for i < len(lines) && strings.HasPrefix(lines[i], "> ") {
				ql = append(ql, lines[i][2:])
				i++
			}
			i--
			qText := renderInline(strings.Join(ql, " "))
			out.WriteString(fmt.Sprintf("<blockquote>%s</blockquote>\n", qText))
			continue
		}

		// Unordered list
		ulMatch := regexp.MustCompile(`^[\s]*[-*+]\s+`).MatchString(line)
		if ulMatch {
			if !inList || listType != "ul" {
				if inList {
					out.WriteString(closeList(listType))

				out.WriteString("<ul>\n")
				inList = true
				listType = "ul"
			}
			item := renderInline(regexp.MustCompile(`^[\s]*[-*+]\s+`).ReplaceAllString(line, ""))
			out.WriteString(fmt.Sprintf("<li>%s</li>\n", item))
			continue
		}

		// Ordered list
		olMatch := regexp.MustCompile(`^\s*\d+\.\s+`).MatchString(line)
		if olMatch {
			if !inList || listType != "ol" {
				if inList {
					out.WriteString(closeList(listType))

				out.WriteString("<ol>\n")
				inList = true
				listType = "ol"
			}
			item := renderInline(regexp.MustCompile(`^\s*\d+\.\s+`).ReplaceAllString(line, ""))
			out.WriteString(fmt.Sprintf("<li>%s</li>\n", item))
			continue
		}

		// If we are in a list and now a non-list line, close list
		if inList && !ulMatch && !olMatch && line != "" {
			out.WriteString(closeList(listType))
			inList = false
			listType = ""
			_ = listStartIndex
			_ = hasList
		}

		// Blank line: close any paragraph
		if strings.TrimSpace(line) == "" {
			continue // blank lines are separators
		}

		// Otherwise: paragraph
		content := renderInline(line)
		out.WriteString(fmt.Sprintf("<p>%s</p>\n", content))

	if inCodeBlock {
		out.WriteString("</code></pre>\n")

	if inList {
		out.WriteString(closeList(listType))

	return out.String()
}
}
}
}
}
}
}