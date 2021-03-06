package deployments

import (
	admissioncontroller "github.com/capital-prawn/capper"

	"k8s.io/api/admission/v1"
	c "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"context"
	"encoding/json"
	log "k8s.io/klog/v2"
	"strconv"
)

func validateCreate() admissioncontroller.AdmitFunc {
	log.Infof("validate create - deployments")

	return func(r *v1.AdmissionRequest) (*admissioncontroller.Result, error) {

		cm, err := getConfigMap()
		if err != nil {
			return &admissioncontroller.Result{Msg: "Unable to get configmap"}, nil
		}

		dp, err := parsePod(r.Object.Raw)
		if err != nil {
			return &admissioncontroller.Result{Msg: err.Error()}, nil
		}

		for _, namespace := range cm.NamespaceWhitelist {
			if namespace == r.Namespace {
				return &admissioncontroller.Result{Allowed: true, Msg: "Deployment is in a whitelisted namespace, skipping"}, nil
			}
		}

		// Now let's set it to the global cap
		global_cap := c.MustParse(cm.GlobalCap)

		if err != nil {
			return &admissioncontroller.Result{Allowed: false, Msg: "Global CPU cap in config map was not able to be converted to an integer"}, nil
		}

		var t1 int64 = 0
		var t2 int64 = 0

		for _, container := range dp.Spec.Containers {
			if v, ok := cm.ApplicationCaps[container.Name]; ok {
				cpu, err := strconv.ParseInt(v, 10, 64)

				if err != nil {
					return &admissioncontroller.Result{Allowed: false, Msg: "Error converting application cap value to int"}, nil
				}
				if t1 > cpu {

					return &admissioncontroller.Result{Allowed: false, Msg: "Container requested more CPU than application-specific cap allows"}, nil
				}
			} else {
				cpu := container.Resources.Requests.Cpu()
				t1 = cpu.Value()
				if err != nil {
					log.Errorf("Error getting CPU request value: %s", err)
					return &admissioncontroller.Result{Allowed: true}, nil
				}
				log.Infof("CPU value is: %d", t1)
				t2, _ = global_cap.AsInt64()
				log.Infof("Checking container request %s against %s", t1, t2)

				if t1 > t2 {
					return &admissioncontroller.Result{Allowed: false, Msg: "CPU request above global CPU cap"}, nil
				}
			}
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
	log.Infof(result.Data["value"])

	if err != nil {
		return nil, err
	}
	return ccm, nil

}

type CapperConfigMap struct {
	NamespaceWhitelist []string          `json:"namespace_whitelist"`
	ApplicationCaps    map[string]string `json:"cpu_request_caps"`
	GlobalCap          string            `json:"global_cpu_request_cap"`
}
