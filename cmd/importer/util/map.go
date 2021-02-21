package util

import (
	"fmt"
	"strings"
)

func MapQuery(m map[string]string) string {
	query := make([]string, 0)
	for key, value := range m {
		selector := fmt.Sprintf("%s=%s", key, value)
		query = append(query, selector)
	}
	return strings.Join(query, "&")
}
