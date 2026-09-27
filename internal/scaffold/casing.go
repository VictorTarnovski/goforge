package scaffold

import "strings"

// splitWords breaks a snake_case, kebab-case, or already-PascalCase/camelCase
// identifier into its component words.
func splitWords(raw string) []string {
	var words []string
	var current strings.Builder

	flush := func() {
		if current.Len() > 0 {
			words = append(words, current.String())
			current.Reset()
		}
	}

	runes := []rune(raw)
	for i, r := range runes {
		switch {
		case r == '_' || r == '-' || r == ' ':
			flush()
		case r >= 'A' && r <= 'Z' && i > 0 && !isUpperOrDigit(runes[i-1]):
			flush()
			current.WriteRune(r)
		default:
			current.WriteRune(r)
		}
	}
	flush()

	return words
}

func isUpperOrDigit(r rune) bool {
	return (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')
}

// PascalCase turns "favorite_food" or "favorite-food" into "FavoriteFood".
func PascalCase(raw string) string {
	var b strings.Builder
	for _, word := range splitWords(raw) {
		if word == "" {
			continue
		}
		b.WriteString(strings.ToUpper(word[:1]))
		b.WriteString(strings.ToLower(word[1:]))
	}
	return b.String()
}

// CamelCase turns "favorite_food" into "favoriteFood".
func CamelCase(raw string) string {
	pascal := PascalCase(raw)
	if pascal == "" {
		return pascal
	}
	return strings.ToLower(pascal[:1]) + pascal[1:]
}

// SnakeCase turns "FavoriteFood" or "favoriteFood" into "favorite_food".
func SnakeCase(raw string) string {
	words := splitWords(raw)
	for i, word := range words {
		words[i] = strings.ToLower(word)
	}
	return strings.Join(words, "_")
}
