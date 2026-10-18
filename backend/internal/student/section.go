package student

import "strconv"

func Section(reg string) string {
	n, err := strconv.Atoi(reg[len(reg)-2:])
	if err != nil || n%2 == 0 {
		return "B"
	}
	return "A"
}