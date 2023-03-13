package flagx

import (
	"strings"
)

type ArrayFlag []string

func (i *ArrayFlag) String() string {
	return ""
}

func (i *ArrayFlag) Set(value string) error {
	*i = append(*i, value)
	return nil
}

type MapFlag map[string]string

func (m *MapFlag) String() string {
	return "my string representation"
}

func (m *MapFlag) Set(value string) error {
	parts := strings.Split(value, "=")
	(*m)[parts[0]] = parts[1]
	return nil
}
