// Copyright 2021-present Open Networking Foundation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gnmi

import (
	"context"
	"math"
	"sync"

	"google.golang.org/grpc"

	topoapi "github.com/onosproject/onos-api/go/onos/topo"

	"github.com/onosproject/onos-lib-go/pkg/errors"
)

// ConnManager gNMI connection manager
type ConnManager interface {
	Get(ctx context.Context, targetID topoapi.ID) (Conn, error)
	List(ctx context.Context) ([]Conn, error)
	Watch(ctx context.Context, ch chan<- ConnEvent) error
	Connect(ctx context.Context, target *topoapi.Object) (Conn, error)
	Close(ctx context.Context, targetID topoapi.ID) error
}

// NewConnManager creates a new gNMI connection manager
func NewConnManager() ConnManager {
	mgr := &connManager{
		conns:   make(map[topoapi.ID]Conn),
		eventCh: make(chan ConnEvent),
	}
	go mgr.processEvents()
	return mgr
}

type connManager struct {
	conns      map[topoapi.ID]Conn
	connsMu    sync.RWMutex
	watchers   []chan<- ConnEvent
	watchersMu sync.RWMutex
	eventCh    chan ConnEvent
}

// newConn creates a new gNMI connection
func (m *connManager) connect(ctx context.Context, target *topoapi.Object) (Conn, error) {
	m.connsMu.Lock()
	currentConn, ok := m.conns[target.ID]
	if ok {
		m.connsMu.Unlock()
		return nil, errors.NewAlreadyExists("gNMI connection %s already exists for target %s", currentConn.ID(), target.ID)
	}
	m.connsMu.Unlock()
	if target.Type != topoapi.Object_ENTITY {
		return nil, errors.NewInvalid("object is not a topo entity %v+", target)
	}

	typeKindID := string(target.GetEntity().KindID)
	if len(typeKindID) == 0 {
		return nil, errors.NewInvalid("target entity %s must have a 'kindID' to work with onos-config", target.ID)
	}

	destination, err := newDestination(target)
	if err != nil {
		log.Warnf("Failed to create a new target %s", err)
		return nil, err
	}
	log.Infof("Connecting to gNMI target: %+v", destination)
	opts := []grpc.DialOption{
		grpc.WithBlock(),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(math.MaxInt32)),
	}

	// Check if device uses an adapter. If it does, set up a client to adapter (if there aren't one already).
	// Currently "netconf-device" is static, but should be more dynamic (it requires a lot more changes which
	// also could include changes in topo, most likely).
	if typeKindID == "netconf-device" {

	}
	// 	// TODO: Check if adapter connection already exists, if so, don't create a new one.
	// 	if _, ok := m.conns[target.ID]; ok {
	// 		log.Infof("Conn already exists")
	// 	} else {
	// 	}

	// 	adapterName := strings.Split(destination.Addrs[0], ":")[0]
	// 	log.Infof("Adding a gNMI connection %s to adapter \"%s\", for target \"%s\"", adapterName, destination.Target)

	// 	// TODO: create conn object and do whatever the normal case is doing for this connection.
	// }

	gnmiClient, clientConn, err := newGNMIClient(ctx, *destination, opts)
	if err != nil {
		log.Warnf("Failed to connect to the gNMI target %s: %s", destination.Target, err)
		return nil, err
	}

	conn := newConn(
		WithClientConn(clientConn),
		WithClient(gnmiClient),
		WithTargetID(target.ID))

	log.Infof("Adding a gNMI connection %s for the target %s", conn.ID(), target.ID)

	m.connsMu.Lock()
	m.conns[target.ID] = conn
	m.connsMu.Unlock()
	m.eventCh <- ConnEvent{
		Conn:      conn,
		EventType: Connected,
	}
	return conn, nil
}

// Connect connecting to a gNMI target and adding a new gNMI connection
func (m *connManager) Connect(ctx context.Context, target *topoapi.Object) (Conn, error) {
	newConn, err := m.connect(ctx, target)
	if err != nil {
		return nil, err
	}
	return newConn, nil
}

// Get returns a gNMI connection based on a given target ID
func (m *connManager) Get(ctx context.Context, targetID topoapi.ID) (Conn, error) {
	m.connsMu.RLock()
	defer m.connsMu.RUnlock()
	conn, ok := m.conns[targetID]
	if !ok {
		return nil, errors.NewNotFound("gNMI connection for target '%s' not found", targetID)
	}
	return conn, nil
}

// List lists all  gNMI connections
func (m *connManager) List(ctx context.Context) ([]Conn, error) {
	m.connsMu.RLock()
	defer m.connsMu.RUnlock()
	conns := make([]Conn, 0, len(m.conns))
	for _, conn := range m.conns {
		conns = append(conns, conn)
	}
	return conns, nil
}

func (m *connManager) Close(ctx context.Context, targetID topoapi.ID) error {
	m.connsMu.Lock()
	defer m.connsMu.Unlock()
	if conn, ok := m.conns[targetID]; ok {
		err := conn.Close()
		if err != nil {
			return err
		}
		delete(m.conns, targetID)
		m.eventCh <- ConnEvent{
			Conn:      conn,
			EventType: Disconnected,
		}
	}
	return nil
}

// Watch watches gNMI connection changes
func (m *connManager) Watch(ctx context.Context, ch chan<- ConnEvent) error {
	m.watchersMu.Lock()
	m.connsMu.Lock()
	m.watchers = append(m.watchers, ch)
	m.watchersMu.Unlock()

	go func() {
		for _, stream := range m.conns {
			ch <- ConnEvent{
				Conn: stream,
			}
		}
		m.connsMu.Unlock()

		<-ctx.Done()
		m.watchersMu.Lock()
		watchers := make([]chan<- ConnEvent, 0, len(m.watchers)-1)
		for _, watcher := range watchers {
			if watcher != ch {
				watchers = append(watchers, watcher)
			}
		}
		m.watchers = watchers
		m.watchersMu.Unlock()
	}()
	return nil
}

func (m *connManager) processEvents() {
	for conn := range m.eventCh {
		m.processEvent(conn)
	}
}

func (m *connManager) processEvent(connEvent ConnEvent) {
	log.Infof("Notifying gNMI connection: %s", connEvent.Conn.ID())
	m.watchersMu.RLock()
	for _, watcher := range m.watchers {
		watcher <- connEvent
	}
	m.watchersMu.RUnlock()
}

var _ ConnManager = &connManager{}
