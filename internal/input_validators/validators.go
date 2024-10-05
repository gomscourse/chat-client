package input_validators

import "strconv"

func NotEmpty(input string) (bool, string) {
	if input == "" {
		return false, "value can't be empty"
	}

	return true, ""
}

func IsInt(input string) (bool, string) {
	_, err := strconv.Atoi(input)
	if err != nil {
		return false, "valid integer required"
	}

	return true, ""
}
