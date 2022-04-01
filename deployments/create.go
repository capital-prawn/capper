package deployments

import (
	admissioncontroller "github.com/capital-prawn/capper"

	"k8s.io/api/admission/v1beta1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	corev1 "k8s.io/client-go/applyconfigurations/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

		for _, namespace := cm.Data["namespace_whitelist"] {
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

func getConfigMap() (*corev1.ConfigMap, err) {
	config, err := rest.InClusterConfig()
	clientset, err := kubernetes.NewForConfig(config)
	result = &corev1.ConfigMap{}
	cm, err := clientset.Get().Namespace("capper").Resource("configmaps").Name("franks-limit-suggester").Do(context.TODO(), metav1.GetOptions{}).Into(result)
	if err != nil {
		return nil, err
	}
	return cm, nil
	
}