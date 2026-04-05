package privacylens

import "regexp"

var tokenRE = regexp.MustCompile(`\[([A-Z][A-Z0-9_]*)_(\d+)\]`)

func Detokenize(text string, vault SessionVault, sessionID string) string {
	return tokenRE.ReplaceAllStringFunc(text, func(token string) string {
		val, err := vault.Retrieve(sessionID, token)
		if err != nil {
			return token
		}
		return val
	})
}
