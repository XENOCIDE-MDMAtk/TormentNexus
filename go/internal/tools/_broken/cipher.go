package tools

import (
	"context"
	"encoding/base64"
	"strings"
)

// HandleRot13 applies ROT13 cipher to text
func HandleRot13(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	result := rot13(text)
	return ok(result)
}

func rot13(s string) string {
	var sb strings.Builder
	sb.Grow(len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'a' && c <= 'z' {
			sb.WriteByte((c-'a'+13)%26 + 'a')
		} else if c >= 'A' && c <= 'Z' {
			sb.WriteByte((c-'A'+13)%26 + 'A')
		} else {
			sb.WriteByte(c)

	}
	return sb.String()
}

}

// HandleCaesar applies Caesar cipher with configurable shift
func HandleCaesar(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	shift, _ :=getInt(args, "shift")
	result := caesar(text, shift)
	return ok(result)
}

func caesar(s string, shift int) string {
	shift = ((shift % 26) + 26) % 26
	var sb strings.Builder
	sb.Grow(len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'a' && c <= 'z' {
			sb.WriteByte((c-'a'+byte(shift))%26 + 'a')
		} else if c >= 'A' && c <= 'Z' {
			sb.WriteByte((c-'A'+byte(shift))%26 + 'A')
		} else {
			sb.WriteByte(c)

	}
	return sb.String()
}

}

// HandleAtbash applies Atbash cipher (alphabet reversal)
func HandleAtbash(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	result := atbash(text)
	return ok(result)
}

func atbash(s string) string {
	var sb strings.Builder
	sb.Grow(len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'a' && c <= 'z' {
			sb.WriteByte('z' - (c - 'a'))
		} else if c >= 'A' && c <= 'Z' {
			sb.WriteByte('Z' - (c - 'A'))
		} else {
			sb.WriteByte(c)

	}
	return sb.String()
}

}

// HandleBase64Encode encodes text to Base64
func HandleBase64Encode(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	encoded := base64.StdEncoding.EncodeToString([]byte(text))
	return ok(encoded)
}

// HandleBase64Decode decodes Base64 text
func HandleBase64Decode(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	decoded, decodeErr := base64.StdEncoding.DecodeString(text)
	if decodeErr != nil {
		return err("invalid Base64 input: " + decodeErr.Error())
}

	return ok(string(decoded))
}