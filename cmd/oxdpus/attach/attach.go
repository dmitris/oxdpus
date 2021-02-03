/*
 * Copyright (c) Sematext Group, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may
 * not use this file except in compliance with the License. You may obtain
 * a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 */

package attach

import (
	"github.com/sematext/oxdpus/pkg/xdp"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// XDP_FLAGS values copied from github.com/iovisor/gobpf/bcc, module.go
// https://github.com/iovisor/gobpf/blob/fb892541d416e3662d2aab072dba3df7410bec94/bcc/module.go#L59-L66
const (
        XDP_FLAGS_UPDATE_IF_NOEXIST = uint32(1) << iota
        XDP_FLAGS_SKB_MODE
        XDP_FLAGS_DRV_MODE
        XDP_FLAGS_HW_MODE
        XDP_FLAGS_MODES = XDP_FLAGS_SKB_MODE | XDP_FLAGS_DRV_MODE | XDP_FLAGS_HW_MODE
        XDP_FLAGS_MASK  = XDP_FLAGS_UPDATE_IF_NOEXIST | XDP_FLAGS_MODES
)

// NewCommand builds a new attach command.
func NewCommand(logger *logrus.Logger) *cobra.Command {
	var flags uint32 
	cmd := &cobra.Command{
		Use:   "attach",
		Short: "Attaches the XDP program on the specified device",
		Run: func(cmd *cobra.Command, args []string) {
			dev, _ := cmd.Flags().GetString("dev")
			mode, _ := cmd.Flags().GetString("mode")
			if mode == "skb" {
				flags = XDP_FLAGS_SKB_MODE
			}
			hook, err := xdp.NewHook()
			if err != nil {
				logger.Fatal(err)
			}
			if err := hook.AttachWithFlags(dev, flags); err != nil {
				logger.Fatal(err)
			}
			logger.Infof("XDP program successfully attached to %s device", dev)
		},
	}
	return cmd
}
