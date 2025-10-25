package parser

import (
	"testing"
)

func TestNameFormat(t *testing.T) {
	names := [][]string{
		{"id_", "_id", "_ID", "id", "iD", "ID", "Id", "order_id", "orderId", "orderID", "OrderID", "__orderId"},
		{"ip_", "_ip", "ip", "iP", "IP", "Ip", "host_ip", "hostIp", "hostIP", "HostIP"},
		{"url_", "_url", "url", "uRL", "URL", "Url", "blog_url", "blogUrl", "blogURL", "BlogURL"},
		{"_user_name", "user_name", "userName", "UserName"},
		{"_zh_中文", "zh_中文", "中文zh"},
	}

	for _, ns := range names {
		var convertNames []string
		var convertNames2 []string
		var convertNames3 []string
		for _, name := range ns {
			convertNames = append(convertNames, toCamel(name))
			convertNames2 = append(convertNames2, customToCamel(name))
			convertNames3 = append(convertNames3, customToSnake(name))
		}
		t.Log("source:             ", ns)
		t.Log("toCamel:           ", convertNames)
		t.Log("customToCamel:", convertNames2)
		t.Log("customToSnake:", convertNames3)
		println()
	}
}
