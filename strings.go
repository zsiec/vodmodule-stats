package vodmodule_stats

import "strconv"

func mustInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}

	return i
}
