package smbdriver

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

func ToKernelMountOptionFlagsAndEnvVars(mountOpts map[string]interface{}) (string, []string) {
	mountFlags, mountEnvVars := separateFlagsAndEnvVars(mountOpts)

	kernelMountOptions := convertToStringArr(renameMountFlags(mountFlags))
	kernelMountEnvVars := convertToStringArr(renameMountFlags(mountEnvVars))

	return strings.Join(kernelMountOptions, ","), kernelMountEnvVars
}

func convertToStringArr(mountOpts map[string]interface{}) []string {
	paramList := []string{}

	for k, v := range mountOpts {
		switch v.(type) {
		case string:
			if val, err := strconv.ParseInt(v.(string), 10, 16); err == nil {
				paramList = append(paramList, fmt.Sprintf("%s=%d", k, val))
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
	return paramList
}

func separateFlagsAndEnvVars(mountOpts map[string]interface{}) (map[string]interface{}, map[string]interface{}) {
	flagList := make(map[string]interface{})
	envVarList := make(map[string]interface{})

	for k, v := range mountOpts {
		if strings.ToLower(k) == "username" {
			envVarList[k] = v
		} else if strings.ToLower(k) == "password" {
			envVarList[k] = v
		} else {
			flagList[k] = v
		}
	}

	return flagList, envVarList
}

func renameMountFlags(mountOpts map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	for k, v := range mountOpts {
		if strings.ToLower(k) == "username" {
			result["USER"] = v
		} else if strings.ToLower(k) == "password" {
			result["PASSWD"] = v
		} else if strings.ToLower(k) == "version" {
			if v != "" {
				result["vers"] = v
			}
		} else if strings.ToLower(k) == "domain" {
			if v != "" {
				result["domain"] = v
			}
		} else {
			result[k] = v
		}
	}
	return result
}
