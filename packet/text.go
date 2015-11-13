package packet

// ParseKVText parses iSCSI key value data.
func ParseKVText(txt []byte) map[string]string {
	m := make(map[string]string)
	var kv, sep int
	var key string
	for i := 0; i < len(txt); i++ {
		if txt[i] == '=' {
			if key == "" {
				sep = i
				key = string(txt[kv:sep])
			}
			continue
		}
		if txt[i] == 0 && key != "" {
			m[key] = string(txt[sep+1 : i])
			key = ""
			kv = i + 1
		}
	}
	return m
}
