package string_wrapper

var greeceMapping = map[rune]rune{
	'a': 'а',
	'b': 'β',
	'c': 'с',
	'd': 'ɗ',
	'e': 'е',
	'f': 'ƒ',
	'g': 'ġ',
	'h': 'һ',
	'i': 'ï',
	'j': 'ʝ',
	'k': 'κ',
	'l': 'ℓ',
	'm': 'ℳ',
	'n': 'ո',
	'o': 'о',
	'p': 'р',
	'q': 'զ',
	'r': 'ɍ',
	's': 'ş',
	't': 'τ',
	'u': 'υ',
	'v': 'ν',
	'w': 'ω',
	'x': 'ҳ',
	'y': 'у',
	'z': 'ʐ',
}

func ReplaceWithSimilarLetter(str string) string {
	var result []rune
	for _, r := range str {
		if r >= 'a' && r <= 'z' {
			if m, ok := greeceMapping[r]; ok {
				result = append(result, m)
			} else {
				result = append(result, r)
			}
		} else {
			result = append(result, r)
		}
	}
	return string(result)
}
