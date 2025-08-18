package term

// Size returns the number of columns and rows in the terminal using native APIs.
func Size() (width, height uint, err error) {
	return size()
}
