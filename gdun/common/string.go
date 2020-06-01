package common

import (
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

//----------------------------------------------------------------------------
// Encoding and Decoding.

func Prune(s string) string {
	return strings.TrimSpace(s)
}

func Escape(s string) string {
	r := url.QueryEscape(strings.TrimSpace(s))

	r = strings.Replace(r, ":", "%3A", -1)
	r = strings.Replace(r, "_", "%5F", -1)
	r = strings.Replace(r, "+", "%20", -1)

	return r
}

func EscapeForStr64(s string) string {
	r := url.QueryEscape(strings.TrimSpace(s))

	r = strings.Replace(r, "%21", "!", -1)
	r = strings.Replace(r, "%40", "@", -1)

	return r
}

func EscapeForSQL(s string) string {
	r := url.QueryEscape(strings.TrimSpace(s))

	r = strings.Replace(r, "%", "\\%", -1)
	return r
}

func Unescape(s string) string {
	r, err := url.QueryUnescape(s)
	if err != nil {
		return s
	}

	return r
}

func ReplaceForSQL(s string) string {
	r := strings.Replace(s, "\\", "\\\\", -1)
	r = strings.Replace(r, "'", "\\'", -1)
	r = strings.Replace(r, "%", "\\%", -1)

	return r
}

func ReplaceForJSON(s string) string {
	r := strings.Replace(s, "\\", "\\\\", -1)
	r = strings.Replace(r, "\"", "\\\"", -1)
	r = strings.Replace(r, "'", "\\'", -1)
	r = strings.Replace(r, "\r", "\\r", -1)
	r = strings.Replace(r, "\n", "\\n", -1)

	return r
}

func UnescapeForJSON(s string) string {
	r, err := url.QueryUnescape(s)
	if err != nil {
		r = s
	}

	// r = strings.Replace(r, "\\", "\\\\", -1)
	// r = strings.Replace(r, "\"", "\\\"", -1)
	// r = strings.Replace(r, "'", "\\'", -1)
	// r = strings.Replace(r, "\r", "\\r", -1)
	// r = strings.Replace(r, "\n", "\\n", -1)
	return ReplaceForJSON(r)
}

func UnescapeForCSV(s string) string {
	r, err := url.QueryUnescape(s)
	if err != nil {
		r = s
	}

	if strings.Index(r, "\n") >= 0 || strings.Index(r, ",") >= 0 || strings.Index(r, "\"") >= 0 {
		return `"` + strings.Replace(s, `"`, `""`, -1) + `"`
	} else {
		return r
	}
}

//----------------------------------------------------------------------------
// Time.

func GetTimeString() string {
	return strconv.FormatInt(time.Now().Unix(), 10)
}

func GetMillisecondString() string {
	return strconv.FormatInt(time.Now().UnixNano()/1e6, 10)
}

//----------------------------------------------------------------------------
// String List.

func InList(s string, l string) bool {
	if len(l) == 0 || len(s) == 0 {
		return false
	}

	ls := strings.Split(l, ",")
	for i := 0; i < len(ls); i++ {
		if strings.Compare(ls[i], s) == 0 {
			return true
		}
	}

	return false
}

func InIntArray(n int, arr []int) bool {
	for i := 0; i < len(arr); i++ {
		if n == arr[i] {
			return true
		}
	}
	return false
}

func AddToList(s string, l string) (string, bool) {
	if len(l) == 0 {
		return s, true
	}

	if InList(s, l) {
		return l, false
	} else {
		return l + "," + s, true
	}
}

func AddToSortedIntList(n int, l string) (string, bool) {
	s := strconv.Itoa(n)

	if len(l) == 0 {
		return s, true
	}

	arr := StringToIntArray(l)
	sort.Ints(arr)

	r := ``
	first := true
	added := false
	for i := 0; i < len(arr); i++ {
		if arr[i] == n {
			return l, false
		}

		if !added {
			if arr[i] > n {
				if first {
					first = false
				} else {
					r += `,`
				}
				r += s

				added = true
			}
		}

		if first {
			first = false
		} else {
			r += `,`
		}
		r += strconv.Itoa(arr[i])
	}

	if !added {
		r += `,` + s
	}
	return r, true
}

func DeleteFromList(s string, l string) (string, bool) {
	if len(l) == 0 {
		return l, false
	}

	flag := false
	r := ""
	first := true
	ls := strings.Split(l, ",")
	for i := 0; i < len(ls); i++ {
		if strings.Compare(ls[i], s) == 0 {
			flag = true
		} else {
			if first {
				first = false
			} else {
				r += ","
			}
			r += ls[i]
		}
	}

	return r, flag
}

//----------------------------------------------------------------------------
// String Map.

func InStringArray(k string, v string, arr []string) bool {
	if len(k) == 0 {
		return false
	}

	target := k + ":" + v
	for i := 0; i < len(arr); i++ {
		if arr[i] == target {
			return true
		}
	}
	return false
}

func InStringArrayByKey(k string, arr []string) (string, bool) {
	if len(k) == 0 {
		return "", false
	}

	target := k + ":"
	for i := 0; i < len(arr); i++ {
		if len(arr) == 0 {
			continue
		}

		if strings.HasPrefix(arr[i], target) {
			return arr[i], true
		}
	}
	return "", false
}

func InMap(k string, m string) bool {
	if len(m) == 0 || len(k) == 0 {
		return false
	}

	arr := strings.Split(m, ",")
	_, existing := InStringArrayByKey(k, arr)
	return existing
}

func AddResourceToMap(k string, v string, m string) (string, bool) {
	if len(m) == 0 {
		return k + ":" + v, true
	}

	r := ""
	first := true
	existing := false
	marr := strings.Split(m, ",")
	for i := 0; i < len(marr); i++ {
		if len(marr[i]) == 0 {
			continue
		}

		kvs := strings.Split(marr[i], ":")
		if strings.Compare(kvs[0], k) == 0 {
			if len(kvs) == 2 {
				if strings.Compare(kvs[1], v) == 0 {
					return m, false
				}
			}

			// Change the value of the key.
			existing = true
			if first {
				first = false
			} else {
				r += ","
			}
			r += k + ":" + v
		} else {
			if first {
				first = false
			} else {
				r += ","
			}
			r += marr[i]
		}
	}

	if existing {
		return r, true
	}

	return r + "," + k + ":" + v, true
}

func DeleteFromMap(k string, m string) (string, bool) {
	if len(m) == 0 || len(k) == 0 {
		return m, false
	}

	flag := false
	r := ""
	first := true
	marr := strings.Split(m, ",")
	for i := 0; i < len(marr); i++ {
		// Ignore empty items.
		if len(marr[i]) == 0 {
			continue
		}

		// Get the key and value.
		kvs := strings.Split(marr[i], ":")

		// Delete this item.
		if strings.Compare(kvs[0], k) == 0 {
			flag = true
			continue
		}

		// Append others.
		if first {
			first = false
		} else {
			r += ","
		}
		r += marr[i]
	}

	return r, flag
}

//----------------------------------------------------------------------------

func IntArrayToString(arr []int) string {
	if arr == nil {
		return ""
	}

	r := ""
	first := true
	for i := 0; i < len(arr); i++ {
		if first {
			first = false
		} else {
			r += ","
		}

		r += strconv.Itoa(arr[i])
	}

	return r
}

func StringToIntArray(l string) []int {
	if len(l) == 0 {
		return []int{}
	}

	ls := strings.Split(l, ",")
	size := len(ls)

	r := make([]int, size)
	for i := 0; i < size; i++ {
		n, err := strconv.Atoi(ls[i])
		if err != nil {
			return []int{}
		}
		r[i] = n
	}

	return r
}

func StringToStringArray(l string) []string {
	if len(l) == 0 {
		return []string{}
	}

	return strings.Split(l, ",")
}

//----------------------------------------------------------------------------

func IntArrayToJSON(arr []int) string {
	s := `[`
	first := true
	for i := 0; i < len(arr); i++ {
		if first {
			first = false
		} else {
			s += `,`
		}
		s += strconv.Itoa(arr[i])
	}
	s += `]`
	return s
}

func StringArrayToJSON(arr []string) string {
	s := `[`
	first := true
	for i := 0; i < len(arr); i++ {
		tmp := strings.TrimSpace(arr[i])
		if len(tmp) == 0 {
			continue
		}

		if first {
			first = false
		} else {
			s += `,`
		}

		s += `"` + UnescapeForJSON(arr[i]) + `"`
	}
	s += `]`
	return s
}

func StringArrayToString(arr []string) string {
	s := ``

	first := true
	for i := 0; i < len(arr); i++ {
		if len(arr[i]) == 0 {
			continue
		}

		if first {
			first = false
		} else {
			s += `,`
		}
		s += arr[i]
	}

	return s
}

func ResourceArrayToJSON(arr []string, numericID bool) string {
	s := `[`
	first := true
	for i := 0; i < len(arr); i++ {
		if len(arr[i]) == 0 {
			continue
		}

		if first {
			first = false
		} else {
			s += `,`
		}

		k := arr[i]
		v := ""
		pos := strings.Index(arr[i], ":")
		if pos > 0 {
			k = (arr[i])[:pos]
			v = (arr[i])[pos+1:]
		}

		s += `{"` + FIELD_ID + `":`
		if numericID {
			s += k
		} else {
			s += `"` + UnescapeForJSON(k) + `"`
		}
		if len(v) > 0 {
			vs := strings.Split(v, "_")
			n := len(vs)

			s += `,"` + FIELD_NAME + `":"`
			if n >= 1 {
				s += UnescapeForJSON(vs[0])
			}
			s += `"`

			if n >= 3 {
				s += `,"` + FIELD_PREPARATION + `":` + UnescapeForJSON(vs[1]) +
					`,"` + FIELD_NECESSARY + `":` + UnescapeForJSON(vs[2])
			} else {
				s += `,"` + FIELD_PREPARATION + `":1` +
					`,"` + FIELD_NECESSARY + `":1`
			}

			if n >= 6 {
				s += `,"` + FIELD_START_TIME + `":` + UnescapeForJSON(vs[3])
				if vs[3] != "0" {
					s += `000`
				}
				s += `,"` + FIELD_DURATION + `":` + UnescapeForJSON(vs[4]) +
					`,"` + FIELD_COUNT + `":` + UnescapeForJSON(vs[5])
			} else {
				s += `,"` + FIELD_START_TIME + `":0` +
					`,"` + FIELD_DURATION + `":0` +
					`,"` + FIELD_COUNT + `":0`
			}
		} else {
			s += `,"` + FIELD_NAME + `":"",` +
				`"` + FIELD_PREPARATION + `":1,` +
				`"` + FIELD_NECESSARY + `":1`
		}
		s += `}`
	}
	s += `]`
	return s
}

//----------------------------------------------------------------------------

func CombineIntArray(arr1 []int, arr2 []int) []int {
	n1 := len(arr1)
	n2 := len(arr2)

	if n1+n2 == 0 {
		return []int{}
	}

	r := make([]int, n1+n2)
	j := 0

	for i := 0; i < n1; i++ {
		r[j] = arr1[i]
		j++
	}
	for i := 0; i < n2; i++ {
		r[j] = arr2[i]
		j++
	}

	return r
}

//----------------------------------------------------------------------------

func CombineIntArrayNumerically(arr1 []int, arr2 []int) []int {
	n1 := len(arr1)
	n2 := len(arr2)
	if (n1 == 0) && (n2 == 0) {
		return []int{}
	}

	n := n1
	if n2 > n1 {
		n = n2
	}
	r := make([]int, n)

	for i := 0; i < n; i++ {
		r[i] = 0
		if i < n1 {
			r[i] += arr1[i]
		}
		if i < n2 {
			r[i] += arr2[i]
		}
	}

	return r
}

//----------------------------------------------------------------------------
