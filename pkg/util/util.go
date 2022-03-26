package util

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type SyncSteps struct {
	Add    []string
	Delete []string
	Common []string
}

// FindSyncSteps - finds sync steps
func FindSyncSteps(src []string, dest []string) SyncSteps {
	m1 := ToMap(src)
	m2 := ToMap(dest)
	add := map[string]bool{}
	del := map[string]bool{}
	common := map[string]bool{}

	for v1 := range m1 {
		_, ok := m2[v1]
		if ok {
			common[v1] = true
		} else {
			del[v1] = true
		}
	}
	for v2 := range m2 {
		_, ok := common[v2]
		if !ok {
			add[v2] = true
		}
	}

	return SyncSteps{Delete: ToArray(del), Add: ToArray(add), Common: ToArray(common)}
}

func ToMap(arr []string) map[string]bool {
	m := map[string]bool{}
	for _, v := range arr {
		m[v] = true
	}
	return m
}

func ToArray(m map[string]bool) []string {
	arr := make([]string, len(m))
	i := 0
	for v := range m {
		arr[i] = v
		i++
	}
	return arr
}

func Contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

func DefaultString(list ...string) string {
	for _, v := range list {
		if v != "" {
			return v
		}
	}
	return ""
}

// GetEnvString - get env value of first key else default.
func GetEnvString(def string, keys ...string) string {
	for _, key := range keys {
		val := os.Getenv(key)
		if val != "" {
			return val
		}
	}
	return def
}

func MD5(data string) string {
	h := md5.New()
	io.WriteString(h, data)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func UnstructuredObject(apiVersion string, kind string) *unstructured.Unstructured {
	u := &unstructured.Unstructured{}
	u.SetAPIVersion(apiVersion)
	u.SetKind(kind)
	return u
}
