// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package config

import (
	topoapi "github.com/onosproject/onos-api/go/onos/topo"
	gnmiapi "github.com/openconfig/gnmi/proto/gnmi"
	"testing"
	"time"

	gnmiutils "github.com/onosproject/onos-config/test/utils/gnmi"
	"github.com/onosproject/onos-config/test/utils/proto"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/status"
)

// TestModels tests GNMI operation involving unknown or illegal paths
func (s *TestSuite) TestModels(t *testing.T) {
	const (
		unknownPath       = "/system/config/no-such-path"
		ntpPath           = "/system/ntp/state/enable-ntp-auth"
		hostNamePath      = "/system/config/hostname"
		clockTimeZonePath = "/system/clock/config/timezone-name"
	)

	ctx, cancel := gnmiutils.MakeContext()
	defer cancel()

	simulator := gnmiutils.CreateSimulator(ctx, t)
	defer gnmiutils.DeleteSimulator(t, simulator)

	// Wait for config to connect to the target
	ready := gnmiutils.WaitForTargetAvailable(ctx, t, topoapi.ID(simulator.Name()), 1*time.Minute)
	assert.True(t, ready)

	// Data to run the test cases
	testCases := []struct {
		description   string
		path          string
		value         string
		valueType     string
		expectedError string
	}{
		{description: "Unknown path", path: unknownPath, valueType: proto.StringVal, value: "123456", expectedError: "no-such-path"},
		{description: "Read only path", path: ntpPath, valueType: proto.BoolVal, value: "false",
			expectedError: "unable to find exact match for RW model path /system/ntp/state/enable-ntp-auth. 113 paths inspected"},
		{description: "Wrong type", path: clockTimeZonePath, valueType: proto.IntVal, value: "11111", expectedError: "expect string"},
		{description: "Constraint violation", path: hostNamePath, valueType: proto.StringVal, value: "not a host name", expectedError: "does not match regular expression pattern"},
	}

	// Make a GNMI client to use for requests
	gnmiClient := gnmiutils.NewOnosConfigGNMIClientOrFail(ctx, t, gnmiutils.NoRetry)

	// Run the test cases
	for _, testCase := range testCases {
		thisTestCase := testCase
		t.Run(thisTestCase.description,
			func(t *testing.T) {
				description := thisTestCase.description
				path := thisTestCase.path
				value := thisTestCase.value
				valueType := thisTestCase.valueType
				expectedError := thisTestCase.expectedError
				t.Logf("testing %q", description)

				setResult := gnmiutils.GetTargetPathWithValue(simulator.Name(), path, value, valueType)
				var setReq = &gnmiutils.SetRequest{
					Ctx:         ctx,
					Client:      gnmiClient,
					Extensions:  gnmiutils.SyncExtension(t),
					Encoding:    gnmiapi.Encoding_PROTO,
					UpdatePaths: setResult,
				}
				msg, _, err := setReq.Set()
				assert.NotNil(t, err, "Set operation for %s does not generate an error", description)
				assert.Contains(t, status.Convert(err).Message(), expectedError,
					"set operation for %s generates wrong error %s", description, msg)
			})
	}
}
