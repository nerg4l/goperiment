package env

import "os"

// Get retrieves the value of the environment variable named by
// the key. If the variable is present in the environment the
// value (which may be empty) is returned. Otherwise the returned
// value will be the fallback value.
func Get(key, fallback string) string {
	val, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	return val
}
