package helpers

// bool.go are funcs that return a boolean.

// Find returns true if the name is found in the collection of names.
func Find(name string, names ...string) bool {
	for _, n := range names {
		if n == name {
			return true
		}
	}
	return false
}
