package parser

import (
	"strings"

	"github.com/huandu/xstrings"
)

// peculiarNouns 是一个特殊名词的映射表，用于将数据库中的特殊名词转换为 Go 语言中的特殊名词。
var peculiarNouns = map[string]string{
	"ID":    "Id",
	"UID":   "Uid",
	"UUID":  "Uuid",
	"GUID":  "Guid",
	"URI":   "Uri",
	"URL":   "Url",
	"IP":    "Ip",
	"QPS":   "Qps",
	"API":   "Api",
	"ASCII": "Ascii",
	"CPU":   "Cpu",
	"CSS":   "Css",
	"DNS":   "Dns",
	"EOF":   "Eof",
	"HTML":  "Html",
	"HTTP":  "Http",
	"HTTPS": "Https",
	"JSON":  "Json",
	"LHS":   "Lhs",
	"RAM":   "Ram",
	"RHS":   "Rhs",
	"RPC":   "Rpc",
	"SLA":   "Sla",
	"SMTP":  "Smtp",
	"SSH":   "Ssh",
	"TLS":   "Tls",
	"TTL":   "Ttl",
	"UI":    "Ui",
	"UTF8":  "Utf8",
	"VM":    "Vm",
	"XML":   "Xml",
	"XSRF":  "Xsrf",
	"XSS":   "Xss",
}

// toCamel is to convert string to pascal case.
//
// if the string is in peculiarNouns, return the upper case.
// otherwise, iterate the peculiarNouns, if length of the string is greater than the length of the peculiar noun,
// compare the suffix, if found, replace it with the upper case.
func toCamel(s string) string {
	// ToCamelCase 是将由空格、下划线和连字符分隔的单词转换为驼峰式大小写。
	str := xstrings.ToPascalCase(s)

	name := strings.ToUpper(str)
	if _, ok := peculiarNouns[name]; ok {
		return name
	}

	// 示例：
	// - "user_xml" -> "UserXml" -> (处理后缀) -> "UserXML"
	// - "xsrf_token" -> "XsrfToken" -> (无需处理) -> "XsrfToken"
	// - "api_web" -> "ApiWeb" -> (无需处理) -> "ApiWeb"
	l := len(str)
	for k, v := range peculiarNouns {
		nl := len(v)
		if l > nl {
			if str[l-nl:] == v {
				str = str[:l-nl] + k
				break
			}
		}
	}

	if str == "_ID" { // special case for table column ID
		str = "ID"
	}

	return str
}

// firstLetterToLower is function to convert first letter to lower case.
//
// if the string is empty, return empty string.
// if the first letter is not letter, return the string as is.
// otherwise, convert the first letter to lower case.
func firstLetterToLower(str string) string {
	if len(str) == 0 {
		return str
	}

	if (str[0] >= 'A' && str[0] <= 'Z') || (str[0] >= 'a' && str[0] <= 'z') {
		return strings.ToLower(str[:1]) + str[1:]
	}

	return str
}

// customToCamel is custom camel case function.
//
// if the string is in peculiarNouns, convert it to lower case.
// otherwise, convert the first letter to lower case.
func customToCamel(str string) string {
	str = toCamel(str)

	if _, ok := peculiarNouns[str]; ok {
		str = strings.ToLower(str)
	} else {
		str = firstLetterToLower(str)
	}

	return str
}

// customToSnake is custom snake case function.
//
// if the string is in peculiarNouns, convert it to lower case.
// otherwise, iterate the peculiarNouns, if length of the string is greater than the length of the peculiar noun,
// compare the suffix, if found, replace it with the lower case.
func customToSnake(str string) string {
	str = toCamel(str)
	l := len(str)
	for k := range peculiarNouns {
		if str == k {
			str = strings.ToLower(str)
			break
		}

		nl := len(k)
		if l > nl {
			if str[l-nl:] == k {
				str = str[:l-nl] + "_" + strings.ToLower(k)
				break
			}
		}
	}

	str = xstrings.ToSnakeCase(str)
	if strings.HasPrefix(str, "__") {
		str = str[1:]
	}

	return str
}
