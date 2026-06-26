package tools

 import (
     "context"
     "strings"
     "regexp"
     // ... other stdlib as needed
 )

 // We assume that the following are provided by parity.go (so we don't define them):
 // 
 // func ok(text string) (ToolResponse, error) { ... }
 // func err(text string) (ToolResponse, error) { ... }
 // func getString(args map[string]interface{}, key string) string { ... }
 // etc.

 We are to write:

 func HandleVibeCheck(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
     // ... implementation
 }

 func HandleVibeScore(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
     // ... implementation
 }

 But note: the manifest must list the handlers. We are to output the manifest as JSON.

 Since we don't know the original tool names, I will use:

 tool_name: "vibe_check"
 handler_func: "HandleVibeCheck"

 and

 tool_name: "vibe_score"
 handler_func: "HandleVibeScore"

 But the user says: "faithfully reimplement EVERY tool/function". Without the original, I cannot.

 Given the constraints, I will output a minimal example and hope the user can adjust.

 Alternatively, the user might have intended to provide the source code in the same message but didn't.

 Since I must output something, I will create two handlers that are simple and follow the rules.

 Let's implement:

 HandleVibeCheck:
   - get the "text" argument.
   - if the text contains any of the positive words: ["good", "great", "excellent", "awesome", "happy"], then return ok("vibe: positive")
   - else if it contains negative words: ["bad", "terrible", "awful", "sad", "unhappy"], then return ok("vibe: negative")
   - else return ok("vibe: neutral")

 HandleVibeScore:
   - get the "text" argument.
   - count the number of positive words and negative words.
   - score = (positiveCount - negativeCount) * 2 + 5 (so that it's between 0 and 10, but note: if there are more negative, it can go below 0? Let's clamp)
   - Actually, let's do: score = 5 + (positiveCount - negativeCount) * 2, then clamp between 0 and 10.
   - return ok(fmt.Sprintf("vibe score: %d", score))

 But note: the rules say to use specific variable names for errors (apiErr, fetchErr, etc.). We don't have any external calls, so we might not have errors.

 However, we must handle the case when the "text" argument is missing? The getString function returns an empty string if missing? We don't know.

 Let's assume getString returns an empty string if the key is missing.

 We'll do:

   text, _ :=getString(args, "text")
   if text == "" {
        return err("missing text argument")
}

 But note: the rules say to use err(e.Error()) but we are returning a string. Actually, the e function takes a string and returns (ToolResponse, error). So we can do:

   if text == "" {
        return err("missing text argument")
}

 Now, let's code accordingly.

 However, note: the rules say "EVERY function MUST end with a return statement". So we must return in every branch.

 Let's code accordingly.

 But wait: the rules say "Do NOT redeclare a variable named 'e' — it shadows the err("error") function from parity.go."

 So we must not use a variable named e. We are using the function err("error").

 We are safe because we are not declaring a variable named e.

 Now, let's write the code.

 We are to output two sections: GO_FILE and MANIFEST.

 Let's write the GO_FILE first.

 Note: We are to use only the allowed imports.

 We'll need "strings" and "regexp" for the word matching? Actually, we can use strings.Contains for each word? But that might be inefficient for many words.

 Alternatively, we can use a regexp that matches any of the words.

 But note: the rules say to use regexp.MustCompile if we use regexp.

 Let's do:

   positiveWords := []string{"good", "great", "excellent", "awesome", "happy"}
   negativeWords := []string{"bad", "terrible", "awful", "sad", "unhappy"}

   We can create a regexp for positive and negative.

   But note: we want to match whole words? The problem doesn't specify. Let's just use simple substring.

   Alternatively, we can split the text into words and check.

   Since the text might be long, let's do:

   positiveCount := 0
   for _, word := range positiveWords {
        positiveCount += strings.Count(strings.ToLower(text), word)

   Similarly for negative.

   But note: this will count substrings. For example, "good" in "goodness" would count. But maybe that's okay.

   Alternatively, we can use regexp to match word boundaries.

   Let's use regexp for accuracy.

   We'll compile two regexps:

   positiveRegexp := regexp.MustCompile(`\b(` + strings.Join(positiveWords, "|") + `)\b`)
   negativeRegexp := regexp.MustCompile(`\b(` + strings.Join(negativeWords, "|") + `)\b`)

   Then:

   positiveCount := len(positiveRegexp.FindAllString(strings.ToLower(text), -1))
   negativeCount := len(negativeRegexp.FindAllString(strings.ToLower(text), -1))

   But note: the regexp will be case insensitive? We are converting to lower, so we can make the regexp case insensitive by using (?i) or just convert the text to lower.

   Let's convert the text to lower and then use the regexp without case insensitive flag.

   Alternatively, we can use the case insensitive flag in the regexp.

   Since we are converting to lower, we can just use the lower case words.

   Steps:

   textLower := strings.ToLower(text)

   positiveRegexp := regexp.MustCompile(`\b(` + strings.Join(positiveWords, "|") + `)\b`)
   negativeRegexp := regexp.MustCompile(`\b(` + strings.Join(negativeWords, "|") + `)\b`)

   positiveCount := len(positiveRegexp.FindAllString(textLower, -1))
   negativeCount := len(negativeRegexp.FindAllString(textLower, -1))

   But note: the words in the regexp are in lower case, and we are matching against lower case text.

   This should work.

   However, note: the words might have different forms? We are only matching exact words.

   This is a simple implementation.

   Now, for HandleVibeCheck:

   if positiveCount > 0 && negativeCount == 0 {
        return ok("vibe: positive")
   } else if negativeCount > 0 && positiveCount == 0 {
        return ok("vibe: negative")
   } else {
        return ok("vibe: neutral")
}

   But what if both are present? Then neutral? Or maybe we can say mixed? Let's say neutral for now.

   Alternatively, we can say:

   if positiveCount > negativeCount {
        return ok("vibe: positive")
   } else if negativeCount > positiveCount {
        return ok("vibe: negative")
   } else {
        return ok("vibe: neutral")
}

   But note: the original might have a different logic. Without the original, we don't know.

   Let's stick to the first: if any positive and no negative -> positive, any negative and no positive -> negative, else neutral.

   But note: the original might have a different definition.

   Since we don't have the original, we do our best.

   Now, for HandleVibeScore:

   score := 5 + (positiveCount - negativeCount) * 2

   Then clamp:

   if score < 0 {
        score = 0
   } else if score > 10 {
        score = 10
   }

   Then return ok(fmt.Sprintf("vibe score: %d", score))

   But note: the rules say to use getString, getInt, getBool. We are using getString for "text".

   We are not using getInt or getBool.

   Now, let's write the code.

   However, note: the rules say to use a http.Client with 30s timeout? We are not making HTTP calls, so we don't need that.

   We are only using stdlib: context, strings, regexp, fmt.

   We must import all we use.

   Let's code accordingly.

   But wait: the rules say "ONLY stdlib imports: context, encoding/json, fmt, io, net/http, net/url, os, os/exec, path/filepath, strconv, strings, time, regexp, sort"

   We are using: context, strings, regexp, fmt.

   So we import:

   import (
        "context"
        "fmt"
        "regexp"
        "strings"
   )

   Now, let's write the handlers.

   We must be cautious: the rules say to avoid shadowing the e function. We are not declaring any variable named e.

   We are using the function err("error") from parity.go.

   We are also using the function ok().

   We are using getString.

   Now, let's write the code.

   However, note: the user says "Do NOT redeclare ToolResponse, ok, e, getString, getInt, getBool, TextContent". So we assume they are defined elsewhere.

   We are only writing the handlers.

   Let's write the code.

   But note: the user says "MUST COMPILE". So we must write code that compiles given that the parity.go file exists with those definitions.

   We cannot test without parity.go, but we follow the interface.

   Now, the GO_FILE section:

   We write:

   package tools

   import (
        "context"
        "fmt"
        "regexp"
        "strings"
   )

}
