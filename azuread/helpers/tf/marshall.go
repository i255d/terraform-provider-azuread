package tf

func ExpandStringSlice(input []interface{}) []string {
	result := make([]string, 0)
	for _, item := range input {
		result = append(result, item.(string))
	}
	return result
}

func ExpandStringSlicePtr(input []interface{}) *[]string {
	r := ExpandStringSlice(input)
	return &r
}

func FlattenStringSlicePtr(input *[]string) []interface{} {
	result := make([]interface{}, 0)
	if input != nil {
		for _, item := range *input {
			result = append(result, item)
		}
	}
	return result
}
