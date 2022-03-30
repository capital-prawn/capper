package main

// http/server.go
func NewServer(port string) *http.Server {
    // Instances hooks
    podsValidation := pods.NewValidationHook()
    // Routers
    ah := newAdmissionHandler()
    mux := http.NewServeMux()
    mux.Handle("/validate/pods", ah.Serve(podsValidation)) // The path of the webhook for Pod validation.
    return &http.Server{
        Addr:    fmt.Sprintf(":%s", port),
        Handler: mux,
    }
}

// cmd/main.go
func main() {
    // flags
    // ...
    server := http.NewServer(port)
    if err := server.ListenAndServeTLS(tlscert, tlskey); err != nil {
        log.Errorf("Failed to listen and serve: %v", err)
    }
}

func validateCreate() admissioncontroller.AdmitFunc {
	return func(r *v1beta1.AdmissionRequest) (*admissioncontroller.Result, error) {
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
	return func(r *v1beta1.AdmissionRequest) (*admissioncontroller.Result, error) {
		var operations []admissioncontroller.PatchOperation
		// ...
		if pod.Namespace == "special" {
			var containers []v1.Container
			containers = append(containers, pod.Spec.Containers...)
			sideC := v1.Container{
				Name:    "test-sidecar",
				Image:   "busybox:stable",
				Command: []string{"sh", "-c", "while true; do echo 'I am a container injected by mutating webhook'; sleep 2; done"},
			}
			containers = append(containers, sideC)
			operations = append(operations, admissioncontroller.ReplacePatchOperation("/spec/containers", containers))
		}
		
		metadata := map[string]string{"origin": "fromMutation"}
		operations = append(operations, admissioncontroller.AddPatchOperation("/metadata/annotations", metadata))
		return &admissioncontroller.Result{
			Allowed:  true,
			PatchOps: operations,
		}, nil
	}
}