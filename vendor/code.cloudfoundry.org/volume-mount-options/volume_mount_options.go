package volume_mount_options

import (
	"fmt"
	"strconv"
	"strings"
)

type MountOpts map[string]interface{}

func NewMountOpts(userOpts map[string]interface{}, mask MountOptsMask) (MountOpts, error) {
	mountOpts := make(map[string]interface{})
	for k, v := range mask.Defaults {
		mountOpts[k] = v
	}

	errorList := []string{}

	for k, v := range userOpts {
		var canonicalKey string
		var ok bool
		if canonicalKey, ok = mask.KeyPerms[k]; !ok {
			canonicalKey = k
		}

		if inArray(mask.Ignored, canonicalKey) {
			continue
		}

		if inArray(mask.Allowed, canonicalKey) {
			uv := uniformKeyData(canonicalKey, v)
			mountOpts[canonicalKey] = uv
		} else if !mask.SloppyMount {
			errorList = append(errorList, k)
		}
	}

	if len(errorList) > 0 {
		return MountOpts{}, fmt.Errorf("Not allowed options: %s", strings.Join(errorList, ", "))
	}

	for _, k := range mask.Mandatory {
		if _, ok := mountOpts[k]; !ok {
			errorList = append(errorList, k)
		}
	}

	if len(errorList) > 0 {
		return MountOpts{}, fmt.Errorf("Missing mandatory options: %s", strings.Join(errorList, ", "))
	}

	return mountOpts, nil
}

func inArray(list []string, key string) bool {
	for _, k := range list {
		if k == key {
			return true
		}
	}

	return false
}

func uniformKeyData(key string, data interface{}) string {
	switch key {
	case "auto-traverse-mounts":
		return uniformData(data, true)

	case "dircache":
		return uniformData(data, true)

	}

	return uniformData(data, false)
}

func uniformData(data interface{}, boolAsInt bool) string {
	switch data.(type) {
	case int, int8, int16, int32, int64, float32, float64:
		return fmt.Sprintf("%#v", data)

	case string:
		return data.(string)

	case bool:
		if boolAsInt {
			if data.(bool) {
				return "1"
			} else {
				return "0"
			}
		} else {
			return strconv.FormatBool(data.(bool))
		}
	}

	return ""
}
