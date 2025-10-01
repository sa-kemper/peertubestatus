package Response

import (
	"errors"
	"flag"
	"html/template"
	"os"
	"strings"

	"github.com/sa-kemper/peertubestats/i18n"
	"github.com/sa-kemper/peertubestats/internal/LogHelp"
	"golang.org/x/text/language"
)

type InternalContextIndex int

const UtilityIndex InternalContextIndex = iota

type Utility struct {
	Template *template.Template
}

func ParseConfigFromEnvFile() (err error) {

	envBytes, err := os.ReadFile(".env")
	if err != nil {
		if errors.Is(err, os.ErrNotExist) { // expected behaviour, the user may choose not to use an env file.
			return nil
		}
		return err
	}

	flag.VisitAll(func(f *flag.Flag) {
		if f.Value.String() != "" {
		}
		if returnVal := extractValFromEnvBytes(envBytes, f.Name); returnVal != "" {
			err = f.Value.Set(returnVal)
			if err != nil {
				return
			}
		}
	})

	return err
}

func ParseConfigFromEnvironment() (err error) {
	flag.VisitAll(func(f *flag.Flag) {
		if val, found := os.LookupEnv(f.Name); found {
			err = f.Value.Set(val)
			if err != nil {
				return
			}
		}
	})
	return
}

// extractValFromEnvBytes extracts a configuration value from a byte slice representing an environment file.
// It parses the input bytes line by line, handling various config file formats including:
// - Simple key=value pairs
// - Quoted values
// - Inline comments
// - Whitespace variations
//
// The function returns the extracted value for the specified key, or an empty string if:
// - The key is not found
// - The value is empty
// - The line is malformed
//
// Example:
//
//	config := []byte(`foo=bar`)
//	value := extractValFromEnvBytes(config, "foo") // returns "bar"
func extractValFromEnvBytes(bytes []byte, name string) string {
	envString := string(bytes)
	envString = strings.ReplaceAll(envString, "\r", "")
	envString = strings.TrimSpace(envString)
	lines := strings.Split(envString, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		equalIndex := strings.Index(line, "=")
		if equalIndex == -1 {
			continue
		}
		key := strings.TrimSpace(line[:equalIndex])
		if key != name {
			continue
		}

		validLine := strings.TrimSpace(line[equalIndex+1:])
		validLine = strings.TrimSpace(validLine)

		firstQuoteIndex := strings.Index(validLine, "\"")
		lastQuoteIndex := strings.Index(validLine[firstQuoteIndex+1:], "\"") + firstQuoteIndex + 1
		var hasMoreThanOneQuote = firstQuoteIndex != lastQuoteIndex
		hashIndex := strings.Index(validLine, "#")
		if firstQuoteIndex != -1 && lastQuoteIndex != -1 && hasMoreThanOneQuote {
			validLine = validLine[firstQuoteIndex+1 : lastQuoteIndex]
			return validLine
		}
		if !hasMoreThanOneQuote && hashIndex != -1 {
			validLine = strings.TrimSpace(validLine[:hashIndex])
			return validLine
		}

		value := validLine
		return value
	}
	return ""
}

func ParseLanguage(lang string) (resolvedLang string, err error) {
	tag, _, err := language.ParseAcceptLanguage(lang)
	LogHelp.LogOnError("Parsing Accept-Language Http Header failed", map[string]string{"Accept-Language": lang}, err)
	for _, langTag := range tag {
		_, ok := i18n.Languages[langTag.String()]
		if ok {
			resolvedLang = langTag.String()
			err = nil
			break
		}
		err = errors.New("could not find a suitable language")
		resolvedLang = "en"
	}
	return resolvedLang, err
}
