package volume_mount_options

import (
	"strconv"

	"code.cloudfoundry.org/volume-mount-options/utils"
	werror "github.com/pkg/errors"
)

type MountOptsMask struct {
	// set of options that are allowed to be provided by the user
	Allowed []string

	// set of default values that will be used if not otherwise provided
	Defaults map[string]interface{}

	// set of key permutations
	KeyPerms map[string]string

	// set of options that, if provided,  will be silently ignored
	Ignored []string

	// set of options that must be provided
	Mandatory []string

	SloppyMount bool
}

func NewMountOptsMask(allowed []string, defaults map[string]interface{}, keyPerms map[string]string, ignored, mandatory []string) (MountOptsMask, error) {
	mask := MountOptsMask{
		Allowed:   allowed,
		Defaults:  defaults,
		KeyPerms:  keyPerms,
		Ignored:   ignored,
		Mandatory: mandatory,
	}

	if defaults == nil {
		mask.Defaults = make(map[string]interface{})
	}

	if v, ok := defaults["sloppy_mount"]; ok {
		vc := utils.InterfaceToString(v)

		var err error
		mask.SloppyMount, err = strconv.ParseBool(vc)

		if err != nil {
			return MountOptsMask{}, werror.Wrap(err, "Invalid sloppy_mount option")
		}
	}

	return mask, nil
}
