package utils

import "strings"

func ExtractArgs(text string, cmd string) []string {
	rets := make([]string, 0)
	cmdIndex := strings.Index(text, cmd)

	if cmdIndex > 0 {
		subText := text[cmdIndex:]
		argStart := strings.Index(subText, "[")
		argEnd := strings.Index(subText, "]")

		if argStart > 0 && argEnd > 0 && argStart < argEnd {
			argText := subText[argStart+1 : argEnd]
			argTokens := strings.Split(argText, ",")
			for _, token := range argTokens {
				if strings.TrimSpace(token) != "" {
					rets = append(rets, strings.TrimSpace(token))
				}
			}
		}
	}

	return rets
}
