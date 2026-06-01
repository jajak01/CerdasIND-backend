package handler

import "strconv"

func parseNumericID(value string) (int64, bool) {
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, false
	}
	return id, true
}
