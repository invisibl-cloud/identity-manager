package flagx

import (
	"strings"
)

// ArrayFlag implements support for array of string flag
type ArrayFlag []string

// String returns string representation of string array flag
func (a *ArrayFlag) String() string {
	sb := strings.Builder{}
	i := 0
	for _, v := range *a {
		if i != 0 {
			sb.WriteString(",")
		}
		sb.WriteString(v)
		i++
	}
	return sb.String()
}

// Set adds new flag to string array
func (a *ArrayFlag) Set(value string) error {
	*a = append(*a, value)
	return nil
}

// MapFlag implements support for map flag
type MapFlag map[string]string

// String returns string representation of map flag
func (m *MapFlag) String() string {
	sb := strings.Builder{}
	i := 0
	for k, v := range *m {
		if i != 0 {
			sb.WriteString(",")
		}
		sb.WriteString(k)
		sb.WriteString("=")
		sb.WriteString(v)
		i++
	}
	return sb.String()
}

// Set adds new flag to map
func (m *MapFlag) Set(value string) error {
	parts := strings.Split(value, "=")
	(*m)[parts[0]] = parts[1]
	return nil
}
