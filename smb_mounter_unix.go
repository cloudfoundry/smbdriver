// +build linux darwin

package smbdriver

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"code.cloudfoundry.org/dockerdriver"
	"code.cloudfoundry.org/dockerdriver/driverhttp"
	"code.cloudfoundry.org/dockerdriver/invoker"
	"code.cloudfoundry.org/goshims/ioutilshim"
	"code.cloudfoundry.org/goshims/osshim"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/volumedriver"
	vmo "code.cloudfoundry.org/volume-mount-options"
	vmou "code.cloudfoundry.org/volume-mount-options/utils"
)

// smbMounter represent volumedriver.Mounter for SMB
type smbMounter struct {
	invoker invoker.Invoker
	osutil  osshim.Os
	ioutil  ioutilshim.Ioutil
	configMask  vmo.MountOptsMask
}

func safeError(e error) error {
	if e == nil {
		return nil
	}
	return dockerdriver.SafeError{SafeDescription: e.Error()}
}

// NewSmbMounter create SMB mounter
func NewSmbMounter(invoker invoker.Invoker, osutil osshim.Os, ioutil ioutilshim.Ioutil, configMask vmo.MountOptsMask) volumedriver.Mounter {
	return &smbMounter{invoker: invoker, osutil: osutil, ioutil: ioutil, configMask: configMask}
}

// Reference: https://www.samba.org/samba/docs/man/manpages-3/mount.cifs.8.html
// Mount mount SMB folder to a local path
// Azure File Service:
//   required: username, password, vers=3.0
//   optional: uid, gid, file_mode, dir_mode, readonly | ro
// Windows Share Folders:
//   required: username, password | sec
//   optional: uid, gid, file_mode, dir_mode, readonly | ro, domain
func (m *smbMounter) Mount(env dockerdriver.Env, source string, target string, opts map[string]interface{}) error {
	logger := env.Logger().Session("smb-mount")
	logger.Info("start")
	defer logger.Info("end")

	mountOpts, err := vmo.NewMountOpts(opts, m.configMask)
	if err != nil {
		logger.Debug("error-parse-entries", lager.Data{
			"given_source":  source,
			"given_target":  target,
			"given_options": opts,
		})
		return safeError(err)
	}

	mountArgs := []string{
		"-t", "cifs",
		source,
		target,
		"-o", vmou.ToKernelMountOptionString(mountOpts),
		"--verbose",
	}

	logger.Debug("parse-mount", lager.Data{
		"given_source":  source,
		"given_target":  target,
		"given_options": opts,
		"mountArgs":  mountArgs,
	})

	logger.Debug("mount", lager.Data{"params": strings.Join(mountArgs, ",")})
	_, err = m.invoker.Invoke(env, "mount", mountArgs)
	return safeError(err)
}

// Unmount unmount a SMB folder from a local path
func (m *smbMounter) Unmount(env dockerdriver.Env, target string) error {
	logger := env.Logger().Session("smb-umount")
	logger.Info("start")
	defer logger.Info("end")

	_, err := m.invoker.Invoke(env, "umount", []string{"-l", target})

	return safeError(err)
}

// Check check whether a local path is mounted or not
func (m *smbMounter) Check(env dockerdriver.Env, name, mountPoint string) bool {
	logger := env.Logger().Session("smb-check-mountpoint")
	logger.Info("start")
	defer logger.Info("end")

	ctx, cancel := context.WithDeadline(context.TODO(), time.Now().Add(time.Second*5))
	defer cancel()
	env = driverhttp.EnvWithContext(ctx, env)
	_, err := m.invoker.Invoke(env, "mountpoint", []string{"-q", mountPoint})

	if err != nil {
		// Note: Created volumes (with no mounts) will be removed
		//       since VolumeInfo.Mountpoint will be an empty string
		logger.Info(fmt.Sprintf("unable to verify volume %s (%s)", name, err.Error()))
		return false
	}
	return true
}

// Purge delete all files in a local path
func (m *smbMounter) Purge(env dockerdriver.Env, path string) {
	logger := env.Logger().Session("purge")
	logger.Info("start")
	defer logger.Info("end")

	fileInfos, err := m.ioutil.ReadDir(path)
	if err != nil {
		logger.Error("purge-readdir-failed", err, lager.Data{"path": path})
		return
	}

	for _, fileInfo := range fileInfos {
		if fileInfo.IsDir() {
			mountDir := filepath.Join(path, fileInfo.Name())

			_, err = m.invoker.Invoke(env, "umount", []string{"-l", "-f", mountDir})
			if err != nil {
				logger.Error("warning-umount-failed", err)
			}

			logger.Info("unmount-successful", lager.Data{"path": mountDir})

			if err := m.osutil.Remove(mountDir); err != nil {
				logger.Error("purge-cannot-remove-directory", err, lager.Data{"name": mountDir, "path": path})
			}

			logger.Info("remove-directory-successful", lager.Data{"path": mountDir})
		}
	}
}
