package content_util

// SplitLongMessage splitLongMessage 分割长消息，避免超过消息长度限制
func SplitLongMessage(message string, maxLength int) []string {
	if len(message) <= maxLength {
		return []string{message}
	}

	var parts []string
	runes := []rune(message) // 使用rune处理中文字符

	for i := 0; i < len(runes); i += maxLength {
		end := i + maxLength
		if end > len(runes) {
			end = len(runes)
		}

		// 尝试在标点符号处分割，避免在句子中间断开
		if end < len(runes) {
			// 向后查找最近的标点符号
			for j := end; j > i && j < len(runes); j-- {
				if IsPunctuation(runes[j]) {
					end = j + 1
					break
				}
			}
		}

		parts = append(parts, string(runes[i:end]))
		i = end - 1 // 调整循环变量
	}

	return parts
}

// IsPunctuation isPunctuation 判断字符是否是中文或英文标点符号
func IsPunctuation(r rune) bool {
	// 中文标点
	if r == '。' || r == '！' || r == '？' || r == '；' || r == '，' || r == '、' || r == '：' || r == '"' || r == '『' || r == '』' || r == '「' || r == '」' {
		return true
	}

	// 英文标点
	if r == '.' || r == '!' || r == '?' || r == ';' || r == ',' || r == ':' ||
		r == '\n' || r == '\r' {
		return true
	}

	return false
}
