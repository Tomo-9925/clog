package judge

import "strconv"

func isNum(s string) bool {
	if _, err := strconv.Atoi(s); err != nil {
		return false
	}
	return true
}
