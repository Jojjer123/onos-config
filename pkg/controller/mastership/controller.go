// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package mastership

import (
	"context"
	"math/rand"
	"time"

	"github.com/onosproject/onos-lib-go/pkg/errors"

	topoapi "github.com/onosproject/onos-api/go/onos/topo"
	"github.com/onosproject/onos-config/pkg/store/topo"
	"github.com/onosproject/onos-lib-go/pkg/controller"

	"github.com/onosproject/onos-lib-go/pkg/logging"
)

const defaultTimeout = 30 * time.Second

var log = logging.GetLogger("controller", "mastership")

// NewController returns a new mastership controller
func NewController(topo topo.Store) *controller.Controller {
	c := controller.NewController("mastership")
	c.Watch(&TopoWatcher{
		topo: topo,
	})

	c.Reconcile(&Reconciler{
		topo: topo,
	})
	return c
}

// Reconciler is mastership reconciler
type Reconciler struct {
	topo topo.Store
}

// Reconcile reconciles the mastership state for a gnmi target
func (r *Reconciler) Reconcile(id controller.ID) (controller.Result, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	targetID := id.Value.(topoapi.ID)
	log.Infof("Reconciling mastership election for the gNMI target  %s", targetID)
	targetEntity, err := r.topo.Get(ctx, targetID)
	if err != nil {
		if errors.IsNotFound(err) {
			return controller.Result{}, nil
		}
		log.Warnf("Failed to reconcile mastership election for the gNMI target with ID %s: %s", targetEntity.ID, err)
		return controller.Result{}, err
	}

	// List the objects in the topo store
	objects, err := r.topo.List(ctx, &topoapi.Filters{
		KindFilter: &topoapi.Filter{
			Filter: &topoapi.Filter_Equal_{
				Equal_: &topoapi.EqualFilter{
					Value: topoapi.CONTROLS,
				},
			},
		},
	})
	if err != nil {
		log.Warnf("Updating MastershipState for target '%s' failed: %v", targetEntity.GetID(), err)
		return controller.Result{}, err
	}
	targetRelations := make(map[topoapi.ID]topoapi.Object)
	for _, object := range objects {
		if object.GetRelation().TgtEntityID == targetID {
			targetRelations[object.ID] = object
		}
	}

	mastership := &topoapi.MastershipState{}
	_ = targetEntity.GetAspect(mastership)
	if _, ok := targetRelations[topoapi.ID(mastership.NodeId)]; !ok {
		if len(targetRelations) == 0 {
			if mastership.NodeId == "" {
				return controller.Result{}, nil
			}
			log.Infof("Master in term %d resigned for the gNMI target '%s'", mastership.Term, targetEntity.GetID())
			mastership.NodeId = ""
		} else {
			// Select a random master to assign to the gnmi target
			relations := make([]topoapi.Object, 0, len(targetRelations))
			for _, targetRelation := range targetRelations {
				relations = append(relations, targetRelation)
			}
			relation := relations[rand.Intn(len(relations))]

			// Increment the mastership term and assign the selected master
			mastership.Term++
			mastership.NodeId = string(relation.ID)
			log.Infof("Elected new master '%s' in term %d for the gNMI target '%s'", mastership.NodeId, mastership.Term, targetEntity.GetID())
		}

		err = targetEntity.SetAspect(mastership)
		if err != nil {
			log.Warnf("Updating MastershipState for gNMI target '%s' failed: %v", targetEntity.GetID(), err)
			return controller.Result{}, err
		}

		// Update the gNMI target entity
		err = r.topo.Update(ctx, targetEntity)
		if err != nil {
			if !errors.IsNotFound(err) && !errors.IsConflict(err) {
				log.Warnf("Updating MastershipState for gNMI target '%s' failed: %v", targetEntity.GetID(), err)
				return controller.Result{}, err
			}
			return controller.Result{}, nil
		}
		return controller.Result{}, nil
	}
	return controller.Result{}, nil
}
