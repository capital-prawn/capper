package deployments

import (
	admissioncontroller "github.com/capital-prawn/capper"

	"k8s.io/api/admission/v1beta1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	// corev1 "k8s.io/client-go/applyconfigurations/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"encoding/json"
	"context"
)

func validateCreate() admissioncontroller.AdmitFunc {
	return func(r *v1beta1.AdmissionRequest) (*admissioncontroller.Result, error) {


		
		cm, err := getConfigMap()
		if err != nil {
			return &admissioncontroller.Result{Msg: "Unable to get configmap"}, nil
		}

		dp, err := parseDeployment(r.Object.Raw)
		if err != nil {
			return &admissioncontroller.Result{Msg: err.Error()}, nil
		}


		for _, namespace := range cm.NamespaceWhitelist {
			if namespace == dp.Namespace {
				return &admissioncontroller.Result{Msg: "Deployment is in a whitelisted namespace, skipping"}, nil
			}
		}

		if dp.Namespace == "special" {
			return &admissioncontroller.Result{Msg: "You cannot create a deployment in `special` namespace."}, nil
		}

		return &admissioncontroller.Result{Allowed: true}, nil
	}
}

func getConfigMap() (*CapperConfigMap, error) {
	config, err := rest.InClusterConfig()
	clientset, err := kubernetes.NewForConfig(config)
	result, err := clientset.CoreV1().ConfigMaps("capper").Get(context.TODO(), "franks-limit-suggester", metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	ccm := &CapperConfigMap{}
	err = json.Unmarshal([]byte(result.Data["value"]), ccm)
	if err != nil {
		return nil, err
	}
	return ccm, nil
	
}

type CapperConfigMap struct {
	  NamespaceWhitelist []string `json:"namespace_whitelist"`
    ApplicationCaps map[string]string `json:"cpu_request_caps"`
    GlobalCap string `json:"global_cpu_request_cap"`
}