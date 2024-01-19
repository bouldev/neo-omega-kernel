package string_wrapper

var unfilteredMapping = map[rune]rune{
	'0': '?', '1': '†', '2': '‡', '3': '⁎', '4': '⁕',
	'5': '⁑', '6': '⁜', '7': '⁂', '8': '✓', '9': '✕',
	'a': '⌁', 'b': ',', 'c': '_', 'd': '~', 'e': '!',
	'f': '@', 'g': '#', 'h': '♪', 'i': '%', 'j': '^',
	'k': '&', 'l': '*', 'm': '(', 'n': ')', 'o': '-',
	'p': '+', 'q': '=', 'r': '[', 's': ']', 't': '‰',
	'u': ';', 'v': '\'', 'w': '⌀', 'x': '<', 'y': '>',
	'z': '‱',
}

func ReplaceWithUnfilteredLetter(str string) string {
	var result []rune
	for _, r := range str {
		if m, ok := unfilteredMapping[r]; ok {
			result = append(result, m)
		} else {
			result = append(result, r)
		}
	}
	return string(result)
}
