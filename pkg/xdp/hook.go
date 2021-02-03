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

package xdp

import (
	"bytes"
	"errors"
	"fmt"
	libbpf "github.com/iovisor/gobpf/elf"
	"github.com/sematext/oxdpus/pkg/xdp/prog/gen"
	"net"
)

const (
	progName = "xdp/xdp_ip_filter"
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
// Hook provides a set of operations that allow for managing the execution of the XDP program
// including attaching it on the network interface, harvesting various statistics or removing
// the program from the interface.
type Hook struct {
	mod *libbpf.Module
}

// NewHook constructs a new instance of the XDP hook from provided XDP code.
func NewHook() (*Hook, error) {
	mod := libbpf.NewModuleFromReader(bytes.NewReader(LoadXDPBytecode()))
	if mod == nil {
		return nil, errors.New("ELF module is not initialized")
	}
	if err := mod.Load(nil); err != nil {
		return nil, err
	}
	return &Hook{mod: mod}, nil
}

// Attach loads the XDP program to specified interface.
func (h *Hook) Attach(dev string) error {
	// before we proceed with attaching make sure that the
	// provided device (interface) is present on the machine
	ifaces, err := net.Interfaces()
	if err != nil {
		return fmt.Errorf("couldn't obtain the list of interfaces: %v", err)
	}
	ok := false
	for _, i := range ifaces {
		if i.Name == dev {
			ok = true
			break
		}
	}
	if !ok {
		return fmt.Errorf("%s interface is not present. Please run `ip a` to list available interfaces", dev)
	}
	// attempt attach the XDP program
	if err := h.mod.AttachXDPWithFlags(dev, progName, XDP_FLAGS_SKB_MODE); err != nil {
		return fmt.Errorf("couldn't attach XDP program to %s interface in skb mode", dev)
	}
	return nil
}

// Remove unloads the XDP program from the interface.
func (h *Hook) Remove(dev string) error {
	if err := h.mod.RemoveXDP(dev); err != nil {
		return fmt.Errorf("couldn't unload XDP program from %s interface", dev)
	}
	return nil
}

// Close closes the underlying eBPF module by disposing any allocated resources.
func (h *Hook) Close() error {
	h.mod.Close()
	return nil
}

// LoadXDPBytecode loads XDP byte code from
func LoadXDPBytecode() []byte {
	b, err := gen.Asset("xdp.o")
	if err != nil {
		panic(fmt.Sprintf("failed to load XDP bytecode from embedded resource: %v", err))
	}
	return b
}
