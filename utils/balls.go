package utils

import "fmt"

func ValidateBalls(reds []int, blue int) error {
	if len(reds) != 6 {
		return fmt.Errorf("红球必须为 6 个")
	}
	seen := make(map[int]bool, 6)
	for _, ball := range reds {
		if ball < 1 || ball > 33 {
			return fmt.Errorf("红球必须在 1–33 之间")
		}
		if seen[ball] {
			return fmt.Errorf("红球不能重复: %d", ball)
		}
		seen[ball] = true
	}
	if blue < 1 || blue > 16 {
		return fmt.Errorf("蓝球必须在 1–16 之间")
	}
	return nil
}
