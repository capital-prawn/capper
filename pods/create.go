package pods

import (
	"strings"

	admissioncontroller "github.com/capital-prawn/capper"

	v1a "k8s.io/api/admission/v1"
		"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	v1 "k8s.io/api/core/v1"
	c "k8s.io/apimachinery/pkg/api/resource"
)

func validateCreate() admissioncontroller.AdmitFunc {
	return func(r *v1a.AdmissionRequest) (*admissioncontroller.Result, error) {
		pod, err := parsePod(r.Object.Raw)
		if err != nil {
			return &admissioncontroller.Result{Msg: err.Error()}, nil
		}

		for _, c := range pod.Spec.Containers {
			if strings.HasSuffix(c.Image, ":latest") {
				return &admissioncontroller.Result{Msg: "You cannot use the tag 'latest' in a container."}, nil
			}
		}

		return &admissioncontroller.Result{Allowed: true}, nil
	}
}

func mutateCreate() admissioncontroller.AdmitFunc {
	return func(r *v1a.AdmissionRequest) (*admissioncontroller.Result, error) {
		var operations []admissioncontroller.PatchOperation
		pod, err := parsePod(r.Object.Raw)
		if err != nil {
			return &admissioncontroller.Result{Msg: err.Error()}, nil
		}

		cm, err := getConfigMap()
		if err != nil {
			return &admissioncontroller.Result{Msg: "Unable to get configmap"}, nil
		}

		global_cap := c.MustParse(cm.GlobalCap)
		var cappedContainers []v1.Container

		for _, container := range dp.Spec.Containers {
			if v, ok := cm.ApplicationCaps[container.Name]; ok {

				newContainer := copy(container)
				if err != nil {
					return &admissioncontroller.Result{Allowed: false, Msg: "Error converting application cap value to int"}, nil
				}

				if t1 > cpu {			
					newContainer.Resources.Requests = v1.ResourceList{"cpu": cpu, "memory": container.Resources.Requests["memory"]}
					
				}
				
			}
			operations = append(operations, admissioncontroller.ReplacePatchOperation("/spec/containers", ))
		}
		// Very simple logic to inject a new "sidecar" container.
		if pod.Namespace == "special" {
			var containers []v1.Container
			containers = append(containers, pod.Spec.Containers...)
			sideC := v1.Container{
				Name:    "test-sidecar",
				Image:   "busybox:stable",
				Command: []string{"sh", "-c", "while true; do echo 'I am a container injected by mutating webhook'; sleep 2; done"},
			}
			
			containers = append(containers, sideC)
			operations = append(operations, admissioncontroller.ReplacePatchOperation("/spec/containers/", containers))
		}

		// Add a simple annotation using `AddPatchOperation`
		metadata := map[string]string{"origin": "fromMutation"}
		operations = append(operations, admissioncontroller.AddPatchOperation("/metadata/annotations", metadata))
		return &admissioncontroller.Result{
			Allowed:  true,
			PatchOps: operations,
		}, nil
	}
}

func getConfigMapPod() (*CapperConfigMapPod, error) {
	config, err := rest.InClusterConfig()
	clientset, err := kubernetes.NewForConfig(config)
	result, err := clientset.CoreV1().ConfigMaps("capper").Get(context.TODO(), "franks-limit-suggester", metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	ccm := &CapperConfigMap{}
	err = json.Unmarshal([]byte(result.Data["value"]), ccm)
	log.Infof(result.Data["value"])

	if err != nil {
		return nil, err
	}
	return ccm, nil
	
}

type CapperConfigMapPod struct {
	NamespaceWhitelist []string `json:"namespace_whitelist"`
    ApplicationCaps map[string]string `json:"cpu_request_caps"`
    GlobalCap string `json:"global_cpu_request_cap"`
}