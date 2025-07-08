package main

import (
	"encoding/json"
	"errors"
)

// TODO: add favicon
type Status struct {
	Version struct {
		Name     string
		Protocol int32
	}
	EnforcesSecureChat bool
	Description        Description
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

type Description struct {
	raw string
	// TODO: add "clean" version
}

func (d *Description) UnmarshalJSON(b []byte) error {
	var v any
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch v := v.(type) {
	case string:
		d.raw = v
	case map[string]any:
		// TODO: respect key order
		if s, ok := v["text"].(string); ok {
			d.raw = s
		} else {
			return errors.New("invalid description")
		}
		// TODO: implement nested extra objects
		if v, ok := v["extra"].([]any); ok {
			for _, v := range v {
				switch v := v.(type) {
				case string:
					d.raw += v
				case map[string]any:
					if s, ok := v["text"].(string); ok {
						d.raw += s
					}
				}
			}
		}
	default:
		return errors.New("invalid description")
	}
	return nil
}
