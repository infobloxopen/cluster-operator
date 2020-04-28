package controller

import (
	"github.com/infobloxopen/cluster-operator/pkg/controller/cluster"
)

// AddToManagerFuncs is a list of functions to add all Controllers to the Manager
var AddToManagerFuncs []func(cluster.ReconcilerConfig) error

// AddToManager adds all Controllers to the Manager
func AddToManager(m cluster.ReconcilerConfig) error {
	for _, f := range AddToManagerFuncs {
		if err := f(m); err != nil {
			return err
		}
	}
	return nil
}
