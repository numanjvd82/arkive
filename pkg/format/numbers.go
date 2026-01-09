package format

import "strconv"

func Commas(value int64) string {
	if value == 0 {
		return "0"
	}
	negative := value < 0
	if negative {
		value = -value
	}
	raw := strconv.FormatInt(value, 10)
	n := len(raw)
	if n <= 3 {
		if negative {
			return "-" + raw
		}
		return raw
	}

	withCommas := make([]byte, 0, n+(n-1)/3)
	for i, r := range raw {
		if i != 0 && (n-i)%3 == 0 {
			withCommas = append(withCommas, ',')
		}
		withCommas = append(withCommas, byte(r))
	}
	if negative {
		return "-" + string(withCommas)
	}
	return string(withCommas)
}
