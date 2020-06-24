package configure

import (
	"strconv"
	"strings"
)

func addParam(b *strings.Builder, name string, value interface{}) {
	switch value.(type) {
	case string:
		if value == "" {
			return
		}
	case uint:
		if value == 0 {
			return
		}
	}

	b.WriteString(name)
	b.WriteByte('=')

	switch value.(type) {
	case string:
		b.WriteString(value.(string))
		b.WriteByte(' ')
	case uint:
		v := uint64(value.(uint))
		b.WriteString(strconv.FormatUint(v, 10))
		b.WriteByte(' ')
	}
}
