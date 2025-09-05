package lib

import "strconv"

func ConvertStringToInt(intToBecome string) (int, error) {
	conv, convErr := strconv.Atoi(intToBecome)
	if convErr != nil {
		return 0, convErr
	}

	return conv, nil
}

func ConvertStringPointerToBoolPointer(boolToBecome *string) *bool {
	if boolToBecome == nil {
		return nil
	}

	conv := *boolToBecome == "true" || *boolToBecome == "1"
	return &conv
}

func NullOrString(input string) *string {
	if input == "" {
		return nil
	}

	return &input
}
