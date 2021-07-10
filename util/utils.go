package util

import "strings"

// SplitNameOrgDomain 将url拆分成节点名称,组织名称和域名
// 默认以'.'为分割符,分割后第1个元素是节点名称,第二个是组织名,
// 第二个到之后所有的内容组为域名
func SplitNameOrgDomain(url string) (string, string, string) {
	firstDotIndex := strings.Index(url, ".")
	name := url[:firstDotIndex]

	args := strings.Split(url, ".")
	orgName := args[1]
	domain := url[firstDotIndex+1:]
	return name, orgName, domain
}

// Indexes 拓展了strings.Index方法，返回子字符串的全部位置
func Indexes(str string, subStr string) []int {
	var indexes []int
	var lastIndex int
	for {
		if index := strings.Index(str, subStr); index != -1 {
			indexes = append(indexes, index+lastIndex)
			lastIndex += index + 1
			str = str[index+1:]
		} else {
			break
		}
	}
	return indexes
}