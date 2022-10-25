package main

// Given two maps, recursively merge right into left, NEVER replacing any key that already exists in left
func MergeMaps(left, right Resource) Resource {
	for key, rightVal := range right {
		if leftVal, present := left[key]; present {
			//then we don't want to replace it - recurse
			_, ok := leftVal.(Resource)
			if !ok {
				left[key] = rightVal
			} else {
				left[key] = MergeMaps(leftVal.(Resource), rightVal.(Resource))
			}
		} else {
			// key not in left so we can just shove it in
			left[key] = rightVal
		}
	}
	return left
}
