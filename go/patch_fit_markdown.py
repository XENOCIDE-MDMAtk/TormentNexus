import re

def add_fit_markdown():
    with open("internal/sync/linkcrawler.go", "r", encoding="utf-8") as f:
        content = f.read()

    new_func = r"""
func extractVisibleText(input string) string {
	// Strip standard non-visible blocks
	withoutScripts := regexp.MustCompile(`(?is)<script[^>]*>.*?</script>`).ReplaceAllString(input, " ")
	withoutStyles := regexp.MustCompile(`(?is)<style[^>]*>.*?</style>`).ReplaceAllString(withoutScripts, " ")
	withoutNoscript := regexp.MustCompile(`(?is)<noscript[^>]*>.*?</noscript>`).ReplaceAllString(withoutStyles, " ")

	// Phase 1 Fit Markdown: aggressively remove boilerplate nav/header/footer/sidebar elements
	withoutNav := regexp.MustCompile(`(?is)<nav[^>]*>.*?</nav>`).ReplaceAllString(withoutNoscript, " ")
	withoutHeader := regexp.MustCompile(`(?is)<header[^>]*>.*?</header>`).ReplaceAllString(withoutNav, " ")
	withoutFooter := regexp.MustCompile(`(?is)<footer[^>]*>.*?</footer>`).ReplaceAllString(withoutHeader, " ")
	withoutAside := regexp.MustCompile(`(?is)<aside[^>]*>.*?</aside>`).ReplaceAllString(withoutFooter, " ")

	// Strip all remaining HTML tags
	withoutTags := regexp.MustCompile(`(?is)<[^>]+>`).ReplaceAllString(withoutAside, " ")

	// Clean up massive whitespace blocks resulting from tag stripping
	collapsedSpaces := regexp.MustCompile(`\s+`).ReplaceAllString(withoutTags, " ")

	return decodeHTMLWhitespace(collapsedSpaces)
}
"""

    # We use simple string replacement instead of regex replacement to avoid escape sequence issues
    start_idx = content.find("func extractVisibleText(input string) string {")
    end_idx = content.find("}", start_idx) + 1

    if start_idx != -1 and "withoutNav" not in content:
        content = content[:start_idx] + new_func.strip() + content[end_idx:]
        with open("internal/sync/linkcrawler.go", "w", encoding="utf-8") as f:
            f.write(content)

add_fit_markdown()
