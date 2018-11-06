package util

import (
	"strings"
	"unicode"
)

// ToCamelCase converts a camelCase-ish string into a typical JSON camelCase string.
// Example conversion: MyJSONVariable -> myJsonVariable
func ToCamelCase(str string) string {
	if len(str) == 0 {
		return ""
	}

	// Split the string into words
	words := make([]string, 0)
	wordStartIndex := 0
	wordEndIndex := -1
	var previousCharacter rune
	for i, character := range str {
		if i > 0 {
			if unicode.IsLower(previousCharacter) && unicode.IsUpper(character) {
				// previousCharacter is definitely the end of a word
				wordEndIndex = i - 1

				word := str[wordStartIndex:i]
				words = append(words, word)
			} else if unicode.IsUpper(previousCharacter) && unicode.IsLower(character) {
				// previousCharacter is definitely the start of a word
				wordStartIndex = i - 1

				if wordStartIndex-wordEndIndex > 1 {
					// This handles consequent uppercase words, such as acronyms.
					// Example: getBlockDAGInfo
					//                  ^^^
					word := str[wordEndIndex+1 : wordStartIndex]
					words = append(words, word)
				}
			}
		}
		previousCharacter = character
	}
	if unicode.IsUpper(previousCharacter) {
		// This handles consequent uppercase words, such as acronyms, at the end of the string
		// Example: TxID
		//            ^^
		for i := len(str) - 1; i >= 0; i-- {
			if unicode.IsLower(rune(str[i])) {
				break
			}

			wordStartIndex = i
		}
	}
	lastWord := str[wordStartIndex:]
	words = append(words, lastWord)

	// Build a PascalCase string out of the words
	var camelCaseBuilder strings.Builder
	for _, word := range words {
		lowercaseWord := strings.ToLower(word)
		capitalizedWord := strings.ToUpper(string(lowercaseWord[0])) + lowercaseWord[1:]
		camelCaseBuilder.WriteString(capitalizedWord)
	}
	camelCaseString := camelCaseBuilder.String()

	// Un-capitalize the first character to covert PascalCase into camelCase
	return strings.ToLower(string(camelCaseString[0])) + camelCaseString[1:]
}
