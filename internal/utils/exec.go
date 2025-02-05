package utils

// ParseCommandLine splits the raw command line string into arguments
func ParseCommandLine(commandLine string) []string {
	var args []string
	var arg string
	var firstQuote rune
	var escapeNext bool

	for _, c := range commandLine {
		if escapeNext {
			arg += string(c)
			escapeNext = false
			continue
		}

		switch c {
		case ' ':
			if firstQuote > 0 {
				arg += string(c)
			} else if arg != "" {
				args = append(args, arg)
				arg = ""
			}
		case '"', '\'':
			switch {
			case c == firstQuote:
				firstQuote = 0
			case firstQuote == 0:
				firstQuote = c
			default:
				arg += string(c)
			}
		case '\\':
			escapeNext = true
		default:
			arg += string(c)
		}
	}
	if arg != "" {
		args = append(args, arg)
	}
	return args
}
