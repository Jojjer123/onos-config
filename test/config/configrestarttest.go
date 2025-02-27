// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package config

import (
	topoapi "github.com/onosproject/onos-api/go/onos/topo"
	gnmiapi "github.com/openconfig/gnmi/proto/gnmi"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"

	gnmiutils "github.com/onosproject/onos-config/test/utils/gnmi"
	hautils "github.com/onosproject/onos-config/test/utils/ha"
	"github.com/onosproject/onos-config/test/utils/proto"
)

const (
	restartTzValue          = "Europe/Milan"
	restartTzPath           = "/system/clock/config/timezone-name"
	restartLoginBannerPath  = "/system/config/login-banner"
	restartMotdBannerPath   = "/system/config/motd-banner"
	restartLoginBannerValue = "LOGIN BANNER"
	restartMotdBannerValue  = "MOTD BANNER"
)

// TestGetOperationAfterNodeRestart tests a Get operation after restarting the onos-config node
func (s *TestSuite) TestGetOperationAfterNodeRestart(t *testing.T) {
	ctx, cancel := gnmiutils.MakeContext()
	defer cancel()

	// Create a simulated target
	simulator := gnmiutils.CreateSimulator(ctx, t)
	defer gnmiutils.DeleteSimulator(t, simulator)

	// Wait for config to connect to the target
	ready := gnmiutils.WaitForTargetAvailable(ctx, t, topoapi.ID(simulator.Name()), 1*time.Minute)
	assert.True(t, ready)

	// Make a GNMI client to use for onos-config requests
	gnmiClient := gnmiutils.NewOnosConfigGNMIClientOrFail(ctx, t, gnmiutils.WithRetry)

	targetPath := gnmiutils.GetTargetPathWithValue(simulator.Name(), restartTzPath, restartTzValue, proto.StringVal)

	// Set a value using onos-config

	var setReq = &gnmiutils.SetRequest{
		Ctx:         ctx,
		Client:      gnmiClient,
		UpdatePaths: targetPath,
		Extensions:  gnmiutils.SyncExtension(t),
		Encoding:    gnmiapi.Encoding_PROTO,
	}
	setReq.SetOrFail(t)

	// Check that the value was set correctly
	var getReq = &gnmiutils.GetRequest{
		Ctx:        ctx,
		Client:     gnmiClient,
		Paths:      targetPath,
		Extensions: gnmiutils.SyncExtension(t),
		Encoding:   gnmiapi.Encoding_PROTO,
	}
	getReq.CheckValues(t, restartTzValue)

	// Restart onos-config
	configPod := hautils.FindPodWithPrefix(t, "onos-config")
	hautils.CrashPodOrFail(t, configPod)

	// Check that the value was set correctly in the new onos-config instance
	getReq.CheckValues(t, restartTzValue)

	// Check that the value is set on the target
	targetGnmiClient := gnmiutils.NewSimulatorGNMIClientOrFail(ctx, t, simulator)
	var getTargetReq = &gnmiutils.GetRequest{
		Ctx:      ctx,
		Client:   targetGnmiClient,
		Encoding: gnmiapi.Encoding_JSON,
		Paths:    targetPath,
	}
	getTargetReq.CheckValues(t, restartTzValue)
}

// TestSetOperationAfterNodeRestart tests a Set operation after restarting the onos-config node
func (s *TestSuite) TestSetOperationAfterNodeRestart(t *testing.T) {
	ctx, cancel := gnmiutils.MakeContext()
	defer cancel()

	// Create a simulated target
	simulator := gnmiutils.CreateSimulator(ctx, t)
	defer gnmiutils.DeleteSimulator(t, simulator)

	// Make a GNMI client to use for onos-config requests
	gnmiClient := gnmiutils.NewOnosConfigGNMIClientOrFail(ctx, t, gnmiutils.WithRetry)

	tzPath := gnmiutils.GetTargetPathWithValue(simulator.Name(), restartTzPath, restartTzValue, proto.StringVal)
	loginBannerPath := gnmiutils.GetTargetPathWithValue(simulator.Name(), restartLoginBannerPath, restartLoginBannerValue, proto.StringVal)
	motdBannerPath := gnmiutils.GetTargetPathWithValue(simulator.Name(), restartMotdBannerPath, restartMotdBannerValue, proto.StringVal)

	targets := []string{simulator.Name(), simulator.Name()}
	paths := []string{restartLoginBannerPath, restartMotdBannerPath}
	values := []string{restartLoginBannerValue, restartMotdBannerValue}

	bannerPaths := gnmiutils.GetTargetPathsWithValues(targets, paths, values)

	// Restart onos-config
	configPod := hautils.FindPodWithPrefix(t, "onos-config")
	hautils.CrashPodOrFail(t, configPod)

	// Set values using onos-config
	var setReq = &gnmiutils.SetRequest{
		Ctx:        ctx,
		Client:     gnmiClient,
		Extensions: gnmiutils.SyncExtension(t),
		Encoding:   gnmiapi.Encoding_PROTO,
	}
	setReq.UpdatePaths = tzPath
	setReq.SetOrFail(t)
	setReq.UpdatePaths = bannerPaths
	setReq.SetOrFail(t)

	// Check that the values were set correctly
	var getConfigReq = &gnmiutils.GetRequest{
		Ctx:        ctx,
		Client:     gnmiClient,
		Extensions: gnmiutils.SyncExtension(t),
		Encoding:   gnmiapi.Encoding_PROTO,
	}
	getConfigReq.Paths = tzPath
	getConfigReq.CheckValues(t, restartTzValue)
	getConfigReq.Paths = loginBannerPath
	getConfigReq.CheckValues(t, restartLoginBannerValue)
	getConfigReq.Paths = motdBannerPath
	getConfigReq.CheckValues(t, restartMotdBannerValue)

	// Check that the values are set on the target
	targetGnmiClient := gnmiutils.NewSimulatorGNMIClientOrFail(ctx, t, simulator)
	var getTargetReq = &gnmiutils.GetRequest{
		Ctx:      ctx,
		Client:   targetGnmiClient,
		Encoding: gnmiapi.Encoding_JSON,
	}
	getTargetReq.Paths = tzPath
	getTargetReq.CheckValues(t, restartTzValue)
	getTargetReq.Paths = loginBannerPath
	getTargetReq.CheckValues(t, restartLoginBannerValue)
	getTargetReq.Paths = motdBannerPath
	getTargetReq.CheckValues(t, restartMotdBannerValue)

}
