package treekit

func IsInt(ch byte) bool {
	if ch >= '0' && ch <= '9' {
		return true
	}
	return false
}

func IsChar(ch byte) bool {
	if ch >= 'a' && ch <= 'z' {
		return true
	}

	if ch >= 'A' && ch <= 'Z' {
		return true
	}

	return false
}
