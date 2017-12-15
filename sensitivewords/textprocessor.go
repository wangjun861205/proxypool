package sensitivewords

import (
	"strings"
)

var replaceList = []string{"\n", "\r", "\t", " "}

func processText(s string) string {
	for _, rs := range replaceList {
		s = strings.Replace(s, rs, "", -1)
	}
	return s
}
