package common

import (
	"regexp"
	"strings"
	"sync"
)

var (
	regexpsMu sync.Mutex
	regexps   = map[string]*regexp.Regexp{}
)

func ReplaceValue(str string, name string, content string) string {
	regexpsMu.Lock()

	re, ok := regexps[name]
	if !ok {
		re = regexp.MustCompile("{" + name + "\\?([^}]*)}")
		regexps[name] = re
	}

	regexpsMu.Unlock()

	if content == "" {
		str = re.ReplaceAllString(str, "")
	} else {
		str = re.ReplaceAllString(str, "$1")
	}

	return strings.ReplaceAll(str, "{"+name+"}", content)
}
