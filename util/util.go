package util

func GetAbolutePath(filename string) string {
	return ""
}

func SliceIndexByte(b []byte, c byte, start int) int {
	if len(b) <= start {
		return -1
	}
	for ; start < len(b); start++ {
		if b[start] == c {
			return start
		}
	}
	return -1
}
