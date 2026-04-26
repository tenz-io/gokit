package cache

func errorMsg(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

// errCode returns the error code from the error.
func errCode(err error) int {
	if err == nil {
		return 0
	}
	return 1
}
