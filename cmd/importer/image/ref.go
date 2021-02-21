package image

import (
	"fmt"
	"strings"
)

type ImageID struct {
	Registry string
	Name     string
	Ref      string
}

func ParseID(s string) (*ImageID, error) {
	s = strings.TrimPrefix(s, "docker-pullable://")
	result := &ImageID{}
	parts := splitReverseN(s, "@", 2)
	if len(parts) < 2 {
		parts = splitReverseN(s, ":", 2)
	}
	if len(parts) < 2 {
		return nil, fmt.Errorf("Invalid")
	}
	result.Ref = parts[1]
	result.Name = parts[0]
	if strings.Count(result.Name, "/") > 1 {
		parts = strings.SplitN(result.Name, "/", 2)
		result.Registry = parts[0]
		result.Name = parts[1]
	}
	return result, nil
}

func (id *ImageID) String() string {
	return fmt.Sprintf(`ImageID{Registry:"%s",Name:"%s",Ref:"%s"}`, id.Registry, id.Name, id.Ref)
}

func splitReverseN(s, sep string, n int) []string {
	switch {
	case n == 0:
		return nil
	case sep == "":
		fallthrough
	case len(s) == 0:
		return []string{s}
	case n < 0:
		n = strings.Count(s, sep) + 1
	}
	a := make([]string, min(len(s), n))
	i := n - 1
	for i > 0 {
		m := strings.LastIndex(s, sep)
		if m < 0 {
			break
		}
		a[i] = s[m+len(sep):]
		s = s[:m]
		i--
	}
	a[i] = s
	return a
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
