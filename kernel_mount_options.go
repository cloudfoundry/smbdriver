package smbdriver

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

func ToKernelMountOptionString(mountOpts map[string]interface{}) string {
	paramList := []string{}
	for k, v := range mountOpts {
		switch v.(type) {
		case string:
			if val, err := strconv.ParseInt(v.(string), 10, 16); err == nil {
				paramList = append(paramList, fmt.Sprintf("%s=%d", k, val))
			} else if strings.ToLower(k) == "domain" && v == "" {
				continue
			} else if strings.ToLower(k) == "version" {
				if v != "" {
					paramList = append(paramList, fmt.Sprintf("vers=%s", v))
				}
			} else if v == "" {
				paramList = append(paramList, k)
			} else {
				paramList = append(paramList, fmt.Sprintf("%s=%s", k, v))
			}
		case int, int8, int16, int32, int64:
			paramList = append(paramList, fmt.Sprintf("%s=%d", k, v))
		case bool:
			paramList = append(paramList, fmt.Sprintf("%s=%t", k, v))
		}
	}

	sort.Strings(paramList)
	return strings.Join(paramList, ",")
}
