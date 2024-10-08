// Copyright (c) Tailscale Inc & AUTHORS
// SPDX-License-Identifier: BSD-3-Clause

// Package syspolicy provides functions to retrieve system settings of a device.
package syspolicy

import (
	"errors"
	"time"

	"tailscale.com/util/syspolicy/internal/loggerx"
	"tailscale.com/util/syspolicy/setting"
)

func GetString(key Key, defaultValue string) (string, error) {
	markHandlerInUse()
	v, err := handler.ReadString(string(key))
	if errors.Is(err, ErrNoSuchKey) {
		return defaultValue, nil
	}
	return v, err
}

func GetUint64(key Key, defaultValue uint64) (uint64, error) {
	markHandlerInUse()
	v, err := handler.ReadUInt64(string(key))
	if errors.Is(err, ErrNoSuchKey) {
		return defaultValue, nil
	}
	return v, err
}

func GetBoolean(key Key, defaultValue bool) (bool, error) {
	markHandlerInUse()
	v, err := handler.ReadBoolean(string(key))
	if errors.Is(err, ErrNoSuchKey) {
		return defaultValue, nil
	}
	return v, err
}

func GetStringArray(key Key, defaultValue []string) ([]string, error) {
	markHandlerInUse()
	v, err := handler.ReadStringArray(string(key))
	if errors.Is(err, ErrNoSuchKey) {
		return defaultValue, nil
	}
	return v, err
}

// GetPreferenceOption loads a policy from the registry that can be
// managed by an enterprise policy management system and allows administrative
// overrides of users' choices in a way that we do not want tailcontrol to have
// the authority to set. It describes user-decides/always/never options, where
// "always" and "never" remove the user's ability to make a selection. If not
// present or set to a different value, "user-decides" is the default.
func GetPreferenceOption(name Key) (setting.PreferenceOption, error) {
	s, err := GetString(name, "user-decides")
	if err != nil {
		return setting.ShowChoiceByPolicy, err
	}
	var opt setting.PreferenceOption
	err = opt.UnmarshalText([]byte(s))
	return opt, err
}

// GetVisibility loads a policy from the registry that can be managed
// by an enterprise policy management system and describes show/hide decisions
// for UI elements. The registry value should be a string set to "show" (return
// true) or "hide" (return true). If not present or set to a different value,
// "show" (return false) is the default.
func GetVisibility(name Key) (setting.Visibility, error) {
	s, err := GetString(name, "show")
	if err != nil {
		return setting.VisibleByPolicy, err
	}
	var visibility setting.Visibility
	visibility.UnmarshalText([]byte(s))
	return visibility, nil
}

// GetDuration loads a policy from the registry that can be managed
// by an enterprise policy management system and describes a duration for some
// action. The registry value should be a string that time.ParseDuration
// understands. If the registry value is "" or can not be processed,
// defaultValue is returned instead.
func GetDuration(name Key, defaultValue time.Duration) (time.Duration, error) {
	opt, err := GetString(name, "")
	if opt == "" || err != nil {
		return defaultValue, err
	}
	v, err := time.ParseDuration(opt)
	if err != nil || v < 0 {
		return defaultValue, nil
	}
	return v, nil
}

// SelectControlURL returns the ControlURL to use based on a value in
// the registry (LoginURL) and the one on disk (in the GUI's
// prefs.conf). If both are empty, it returns a default value. (It
// always return a non-empty value)
//
// See https://github.com/tailscale/tailscale/issues/2798 for some background.
func SelectControlURL(reg, disk string) string {
	const def = "https://controlplane.tailscale.com"

	// Prior to Dec 2020's commit 739b02e6, the installer
	// wrote a LoginURL value of https://login.tailscale.com to the registry.
	const oldRegDef = "https://login.tailscale.com"

	// If they have an explicit value in the registry, use it,
	// unless it's an old default value from an old installer.
	// Then we have to see which is better.
	if reg != "" {
		if reg != oldRegDef {
			// Something explicit in the registry that we didn't
			// set ourselves by the installer.
			return reg
		}
		if disk == "" {
			// Something in the registry is better than nothing on disk.
			return reg
		}
		if disk != def && disk != oldRegDef {
			// The value in the registry is the old
			// default (login.tailscale.com) but the value
			// on disk is neither our old nor new default
			// value, so it must be some custom thing that
			// the user cares about. Prefer the disk value.
			return disk
		}
	}
	if disk != "" {
		return disk
	}
	return def
}

// SetDebugLoggingEnabled controls whether spammy debug logging is enabled.
func SetDebugLoggingEnabled(v bool) {
	loggerx.SetDebugLoggingEnabled(v)
}
