package main

// Given two maps, recursively merge right into left, replacing any key that already exists in left.
// Modified from: https://stackoverflow.com/a/60420264/1378682
func MergeMaps(left, right Resource) Resource {
	for key, rightVal := range right {
		if leftVal, present := left[key]; present {
			_, ok := leftVal.(Resource)
			if !ok {
				// not a Resource, overwrite value
				left[key] = rightVal
			} else {
				// it's a Resource, merge values
				left[key] = MergeMaps(leftVal.(Resource), rightVal.(Resource))
			}
		} else {
			// add new key
			left[key] = rightVal
		}
	}
	return left
}
