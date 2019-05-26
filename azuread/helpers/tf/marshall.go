package tf

func ExpandStringArray(input []interface{}) []string {
	result := make([]string, 0)
	for _, item := range input {
		result = append(result, item.(string))
	}
	return result
}

func ExpandStringArrayPtr(input []interface{}) *[]string {
	r := ExpandStringArray(input)
	return &r
}

func FlattenStringArrayPtr(input *[]string) []interface{} {
	result := make([]interface{}, 0)
	if input != nil {
		for _, item := range *input {
			result = append(result, item)
		}
	}
	return result
}