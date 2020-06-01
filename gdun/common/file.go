package common

//----------------------------------------------------------------------------

func RemoveFilenameSuffix(filename string) string {
	n := len(filename)
	if n == 0 {
		return ""
	}

	for i := n - 1; i >= 0; i-- {
		if filename[i] == '.' {
			return filename[0:i]
		}
	}
	return filename
}

//----------------------------------------------------------------------------
