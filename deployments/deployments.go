package deployments

import (
	"encoding/json"

	admissioncontroller "github.com/capital-prawn/capper"

	v1 "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
)

// NewValidationHook creates a new instance of deployment validation hook
func NewValidationHook() admissioncontroller.Hook {
	return admissioncontroller.Hook{
		Create: validateCreate(),
		Delete: validateDelete(),
	}
}

func parseDeployment(object []byte) (*v1.Deployment, error) {
	var dp v1.Deployment
	if err := json.Unmarshal(object, &dp); err != nil {
		return nil, err
	}

	return &dp, nil
}

func parsePod(object []byte) (*core.Pod, error) {
	var p core.Pod
	if err := json.Unmarshal(object, &p); err != nil {
		return nil, err
	}

	return &p, nil
}
