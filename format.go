package logger

// formatValue formats a value for serialization
import (
	"bytes"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	timeFormat     = "2006-01-02T15:04:05-0700"
	timeTermFormat = "2006-01-02 15:04"
	floatFormat    = 'f'
)

func formatTerminal(reqData map[string]interface{}, userData map[string]interface{}) string {
	var stringArray []string
	// Choose the color that suits our status.
	var statusColor []byte
	switch {
	case reqData["status"].(int) < 200:
		statusColor = bBlue
	case reqData["status"].(int) < 300:
		statusColor = bGreen
	case reqData["status"].(int) < 400:
		statusColor = bCyan
	case reqData["status"].(int) < 500:
		statusColor = bYellow
	default:
		statusColor = bRed
	}
	reqString := fmt.Sprintf("%s %s %s %s - %dB | %v",
		reqData["method"], reqData["url"], cW(formatValue(reqData["status"]), statusColor),
		reqData["duration"], reqData["size"], time.Now().Format(timeTermFormat))

	stringArray = append(stringArray, reqString)
	stringArray = append(stringArray, "\n")

	for k, v := range userData {
		s := fmt.Sprintf("%v:%v", cW(k, bMagenta), formatValue(v))
		stringArray = append(stringArray, s)
	}
	// Join them into one string
	return strings.Join(stringArray, " ")
}

func formatSyslog(reqData map[string]interface{}, userData map[string]interface{}) []byte {
	var stringArray []string
	reqString := fmt.Sprintf("%s %s %s %s - %dB | %v",
		reqData["method"], reqData["url"], formatValue(reqData["status"]),
		reqData["duration"], reqData["size"], time.Now().Format(timeFormat))

	stringArray = append(stringArray, reqString)

	for k, v := range userData {
		s := fmt.Sprintf("%v:%v", k, formatValue(v))
		stringArray = append(stringArray, s)
	}
	// Join them into one string
	justString := strings.Join(stringArray, " ")
	return []byte(justString)
}

func formatValue(value interface{}) string {
	if value == nil {
		return "nil"
	}

	if t, ok := value.(time.Time); ok {
		// Performance optimization: No need for escaping since the provided
		// timeFormat doesn't have any escape characters, and escaping is
		// expensive.
		return t.Format(timeFormat)
	}
	v := reflect.ValueOf(value)
	if v.Kind() == reflect.Ptr && v.IsNil() {
		return "nil"
	}
	switch v := value.(type) {
	case bool:
		return strconv.FormatBool(v)
	case float32:
		return strconv.FormatFloat(float64(v), floatFormat, 3, 64)
	case float64:
		return strconv.FormatFloat(v, floatFormat, 3, 64)
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", value)
	case string:
		return escapeString(v)
	case error:
		return escapeString(v.Error())
	default:
		return escapeString(fmt.Sprintf("%v", value))
	}
}

var stringBufPool = sync.Pool{
	New: func() interface{} { return new(bytes.Buffer) },
}

// based on
func escapeString(s string) string {
	needsQuotes := false
	needsEscape := false
	for _, r := range s {
		if r <= ' ' || r == '=' || r == '"' || r == ':' {
			needsQuotes = true
		}
		if r == '\\' || r == '"' || r == '\n' || r == '\r' || r == '\t' {
			needsEscape = true
		}
	}
	if needsEscape == false && needsQuotes == false {
		return s
	}
	e := stringBufPool.Get().(*bytes.Buffer)
	e.WriteByte('"')
	for _, r := range s {
		switch r {
		case '\\', '"':
			e.WriteByte('\\')
			e.WriteByte(byte(r))
		case '\n':
			e.WriteString("\\n")
		case '\r':
			e.WriteString("\\r")
		case '\t':
			e.WriteString("\\t")
		default:
			e.WriteRune(r)
		}
	}
	e.WriteByte('"')
	var ret string
	if needsQuotes {
		ret = e.String()
	} else {
		ret = string(e.Bytes()[1 : e.Len()-1])
	}
	e.Reset()
	stringBufPool.Put(e)
	return ret
}
