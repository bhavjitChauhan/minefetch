package main

// TODO: add favicon
type status struct {
	Version struct {
		Name     string
		Protocol int32
	}
	EnforcesSecureChat bool
	Description        text
	Players            struct {
		Max    int
		Online int
		Sample []struct {
			Id   string
			Name string
		}
	}
	Favicon string
}
