package controller

import (
	"github.com/submariner-operator/submariner-operator/pkg/controller/routeagent"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, routeagent.Add)
}
