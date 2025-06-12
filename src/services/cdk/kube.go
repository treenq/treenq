package cdk

import (
	"context"
	"crypto/sha256"
	"encoding/base64" // Make sure this is present
	"encoding/hex"
	"fmt"
	// "path/filepath" // No longer needed after switching to image.FullPath()
	"strings"
	// "sync" // No longer needed due to DefineApp refactor

	tqsdk "github.com/treenq/treenq/pkg/sdk"
	"github.com/treenq/treenq/src/domain"

	appsv1 "k8s.io/api/apps/v1" // Added
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1" // Added
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource" // Added
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/util/intstr" // Added
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	kyaml "sigs.k8s.io/yaml" // Added for marshalling Kubernetes objects
)

const (
	defaultReplicas = 1 // Added constant
)

type Kube struct {
	// host holds a main app host, used to create sub hosts for a quick app preview
	host           string
	dockerRegistry string
	userName       string
	userPassword   string

	// jsii lock (mx) removed as cdk8s is no longer used directly in DefineApp
}

func NewKube(
	host, dockerRegistry, userName, userPassword string,
) *Kube {
	return &Kube{host: host, dockerRegistry: dockerRegistry, userName: userName, userPassword: userPassword}
}

// generateCdk8sStyleSuffix computes a SHA256 hash of the input string,
// takes the first 8 hex characters, and returns it.
func generateCdk8sStyleSuffix(input string) string {
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])[:8]
}

// generateCdk8sStyleName returns a name formatted as chartID-constructID-suffix.
func generateCdk8sStyleName(chartID, constructID string) string {
	return fmt.Sprintf("%s-%s-%s", chartID, constructID, generateCdk8sStyleSuffix(chartID+"/"+constructID))
}

// generateCdk8sStyleAddr returns an address label formatted as chartID-constructID-suffix.
func generateCdk8sStyleAddr(chartID, constructID string) string {
	return fmt.Sprintf("%s-%s-%s", chartID, constructID, generateCdk8sStyleSuffix(chartID+"/"+constructID+"/addr"))
}

// DefineApp generates a Kubernetes manifest string for an application.
// It calls generateKubeResources to create Kubernetes objects and then serializes them to YAML.
// The ctx parameter is currently unused but kept for potential future use (e.g. logging, cancellation).
func (k *Kube) DefineApp(_ context.Context, id string, nsName string, app tqsdk.Space, image domain.Image, secretKeys []string) (string, error) {
	// Lock (k.mx) removed as cdk8s and jsii are no longer directly used here.

	resources, err := k.generateKubeResources(id, nsName, app, image, secretKeys)
	if err != nil {
		return "", fmt.Errorf("failed to generate Kubernetes resources: %w", err)
	}

	var finalYamlElements []string
	for _, res := range resources {
		yamlBytes, err := kyaml.Marshal(res)
		if err != nil {
			return "", fmt.Errorf("failed to marshal resource to YAML: %w", err)
		}
		finalYamlElements = append(finalYamlElements, string(yamlBytes))
	}

	return strings.Join(finalYamlElements, "---\n"), nil
}

// ns generates the namespace name. Kept for other functions that might use it.
func ns(space, repoID string) string {
	return space + "-" + repoID
}

// secretName generates the secret object name. Kept for other functions that might use it.
func secretName(repoID, key string) string {
	return repoID + "-" + strings.ToLower(key)
}

// generateKubeResources creates the Kubernetes resource objects for an application.
// It replaces the logic previously in newAppChart.
func (k *Kube) generateKubeResources(appID, chartID string, appCfg tqsdk.Space, image domain.Image, secretKeys []string) ([]interface{}, error) {
	// Namespace name construction based on previous logic: ns(chartID, appID)
	// chartID was 'nsName' from DefineApp, appID was 'id' from DefineApp.
	// Example: if chartID (nsName)="space", appID (id)="simple-app", then fullNsName="space-simple-app"
	fullNsName := ns(chartID, appID)

	// 1. Namespace
	namespace := &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "Namespace"},
		ObjectMeta: metav1.ObjectMeta{
			Name: fullNsName,
		},
	}

	// 2. Registry Secret
	registrySecretName := "registry-credentials"
	authStr := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", k.userName, k.userPassword)))
	dockerConfigJSON := fmt.Sprintf(`{"auths":{"%s":{"auth":"%s"}}}`, k.dockerRegistry, authStr)
	// The test data has the JSON string as a multi-line block scalar.
	// We'll create the string directly. The exact YAML formatting (multi-line vs single-line)
	// will depend on the marshaller, but the content should be equivalent.
	// To achieve the exact multi-line pretty print from testdata for the json string itself is complex via struct marshalling.
	// The key is that the *content* of .dockerconfigjson is a valid JSON string as expected.
	// Let's ensure the JSON string itself is formatted to be more like the test data's internal structure.
	// The test data's .dockerconfigjson is:
	// {
	//     "auths": {
	//         "registry:5000": {
	//             "auth": "dGVzdHVzZXI6dGVzdHBhc3N3b3Jk"
	//         }
	//     }
	// }
	// This can be achieved by being more literal with the JSON string construction if needed,
	// but for now, the compact JSON string is functionally identical.
	// The test data YAML shows the *string value* of .dockerconfigjson as a block scalar.
	// My current dockerConfigJSON is a compact JSON. The test's expected is:
	// `{\n    \t\t                \"auths\": {\n    \t\t                    \"registry:5000\": {\n    \t\t                        \"auth\": \"dGVzdHVzZXI6dGVzdHBhc3N3b3Jk\"\n    \t\t                    }\n    \t\t                }\n    \t\t            }`
	// This specific formatting of the JSON string itself is hard to replicate perfectly without manual string building.
	// The important part is `auth` field is used and base64 encoded.
	immutableFalse := false
	registrySecret := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "Secret"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      registrySecretName,
			Namespace: fullNsName,
		},
		Immutable: &immutableFalse, // To match test data explicit false
		StringData: map[string]string{
			".dockerconfigjson": dockerConfigJSON,
		},
		Type: corev1.SecretTypeDockerConfigJson, // This type implies K8s expects the content of .dockerconfigjson to be the JSON map.
	}

	// 3. Deployment
	// Construct IDs for naming are based on appCfg.Service.Name, similar to cdk8s construct IDs.
	deploymentConstructID := appCfg.Service.Name + "-deployment"
	deploymentName := generateCdk8sStyleName(chartID, deploymentConstructID)
	deploymentAddrLabelValue := generateCdk8sStyleAddr(chartID, deploymentConstructID)

	replicas := int32(defaultReplicas)
	if appCfg.Service.Replicas > 0 { // appCfg.Service.Replicas was jsii.Number, now int
		replicas = int32(appCfg.Service.Replicas)
	}

	var envVars []corev1.EnvVar
	for key, value := range appCfg.Service.RuntimeEnvs {
		envVars = append(envVars, corev1.EnvVar{Name: key, Value: value})
	}
	for _, key := range secretKeys {
		// Secret object name for ValueFrom uses the global secretName func: secretName(appID, key)
		// This matches the pattern from problem: `id + "-" + strings.ToLower(key)` -> appID is `id`
		// Example: if appID="simple-app", key="MY_SECRET", secret obj name is "simple-app-my_secret"
		// The problem description for secretKeyRef was: id + "-" + strings.ToLower(key) (e.g., id-1234-secret)
		// Let's use the `secretName` function for consistency with StoreSecret.
		secretObjNameForRef := secretName(appID, key)
		envVars = append(envVars, corev1.EnvVar{
			Name: strings.ToUpper(key), // Match original cdk8s and common practice
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{Name: secretObjNameForRef},
					Key:                  key,
				},
			},
		})
	}

	// Resource computation
	// Original cdk8s code used appCfg.Service.ComputationResource directly.
	// CpuUnits (int) -> millicores (e.g. 100m)
	// MemoryMibs (int) -> Mi (e.g. 128Mi)
	// DiskGibs (int) -> Gi (e.g. 1Gi) - EphemeralStorage, not directly mapped to CPU/Memory requests/limits
	// cpuRequest := resource.MustParse("0") // Replaced by cpuReqStr logic
	// memRequest := resource.MustParse("0") // Replaced by memReqStr logic
	// cpuLimit := resource.MustParse("0") // Original base values, not used directly anymore
	// memLimit := resource.MustParse("0")

	computeRes := appCfg.Service.ComputationResource
	// if computeRes.CpuUnits > 0 { // These are handled by cpuReqStr/LimStr logic below
	// 	cpuRequest = resource.MustParse(fmt.Sprintf("%dm", computeRes.CpuUnits))
	// }
	// if computeRes.MemoryMibs > 0 {
	// 	memRequest = resource.MustParse(fmt.Sprintf("%dMi", computeRes.MemoryMibs))
	// }
	// Limits are same as requests as per original cdk8s setup
	// cpuLimit := cpuRequest // These are now effectively replaced by LimStr variables
	// memLimit := memRequest


	// Security Context values from original cdk8s plus code
	allowPrivilegeEscalation := false
	privileged := false // cdk8splus 'Privileged'
	readOnlyRootFilesystem := true // cdk8splus 'ReadOnlyRootFilesystem'
	runAsNonRoot := true // cdk8splus 'EnsureNonRoot' for container, 'EnsureNonRoot' for pod
	runAsUser := int64(1000) // cdk8splus 'User' for container and pod

	// Other Pod Spec values from testdata/app.yaml or defaults
	automountServiceAccountToken := false
	hostNetwork := false // Not in cdk8s, from app.yaml
	dnsPolicy := corev1.DNSClusterFirst // from app.yaml
	restartPolicy := corev1.RestartPolicyAlways // from app.yaml
	setHostnameAsFQDN := false // from app.yaml
	terminationGracePeriodSeconds := int64(30) // from app.yaml
	// minReadySeconds variable removed, set directly in Deployment Spec below
	progressDeadlineSeconds := int32(600) // from app.yaml
	// revisionHistoryLimit := int32(10) // Default in k8s, not explicitly set
	// Ensure "0m" and "0Mi" for zero resources, and handle ephemeral storage
	cpuReqStr := "0m"
	memReqStr := "0Mi"
	ephemeralStorageReqStr := "0Gi"

	cpuLimStr := "0m" // Expected "0m" when computeRes.CpuUnits is 0
	memLimStr := "0Mi" // Expected "0Mi" when computeRes.MemoryMibs is 0
	ephemeralStorageLimStr := "0Gi" // Expected "0Gi" when computeRes.DiskGibs is 0


	if computeRes.CpuUnits > 0 {
		cpuReqStr = fmt.Sprintf("%dm", computeRes.CpuUnits)
		cpuLimStr = cpuReqStr // Limits are same as requests
	}
	if computeRes.MemoryMibs > 0 {
		memReqStr = fmt.Sprintf("%dMi", computeRes.MemoryMibs)
		memLimStr = memReqStr // Limits are same as requests
	}
	if computeRes.DiskGibs > 0 { // Assuming DiskGibs maps to ephemeral-storage
		ephemeralStorageReqStr = fmt.Sprintf("%dGi", computeRes.DiskGibs)
		ephemeralStorageLimStr = ephemeralStorageReqStr // Limits are same as requests
	}


	deployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{APIVersion: "apps/v1", Kind: "Deployment"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentName,
			Namespace: fullNsName,
			// Labels:    map[string]string{"app": deploymentName}, // Removed top-level app label
		},
		Spec: appsv1.DeploymentSpec{
			Replicas:        int32Ptr(replicas),
			MinReadySeconds: int32(0), // Explicitly set to 0 to match testdata
			ProgressDeadlineSeconds: int32Ptr(progressDeadlineSeconds),
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RollingUpdateDeploymentStrategyType,
				RollingUpdate: &appsv1.RollingUpdateDeployment{
					MaxUnavailable: &intstr.IntOrString{Type: intstr.String, StrVal: "25%"},
					MaxSurge:       &intstr.IntOrString{Type: intstr.String, StrVal: "25%"},
				},
			},
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"cdk8s.io/metadata.addr": deploymentAddrLabelValue}, // Changed label key
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"cdk8s.io/metadata.addr": deploymentAddrLabelValue}, // Changed label key
				},
				Spec: corev1.PodSpec{
					AutomountServiceAccountToken: &automountServiceAccountToken,
					Containers: []corev1.Container{{
						Name:  appCfg.Service.Name,
						Image: image.FullPath(),
						ImagePullPolicy: corev1.PullAlways, // Added ImagePullPolicy
						Ports: []corev1.ContainerPort{{
							Name: "http", // Added port name
							ContainerPort: int32(appCfg.Service.HttpPort),
						}},
						Env:   envVars,
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:              resource.MustParse(cpuReqStr),
								corev1.ResourceMemory:           resource.MustParse(memReqStr),
								corev1.ResourceEphemeralStorage: resource.MustParse(ephemeralStorageReqStr),
							},
							Limits: corev1.ResourceList{
								corev1.ResourceCPU:              resource.MustParse(cpuLimStr),
								corev1.ResourceMemory:           resource.MustParse(memLimStr),
								corev1.ResourceEphemeralStorage: resource.MustParse(ephemeralStorageLimStr),
							},
						},
						SecurityContext: &corev1.SecurityContext{
							AllowPrivilegeEscalation: &allowPrivilegeEscalation,
							Privileged:               &privileged,
							ReadOnlyRootFilesystem:   &readOnlyRootFilesystem,
							RunAsNonRoot:             &runAsNonRoot,
							RunAsUser:                &runAsUser,
						},
					}},
					DNSPolicy:                     dnsPolicy,
					HostNetwork:                   hostNetwork, // Kept from previous logic, testdata has false
					ImagePullSecrets:              []corev1.LocalObjectReference{{Name: registrySecretName}},
					RestartPolicy:                 restartPolicy,
					SecurityContext: &corev1.PodSecurityContext{
						RunAsUser:    &runAsUser,
						RunAsNonRoot: &runAsNonRoot,
						FSGroupChangePolicy: func() *corev1.PodFSGroupChangePolicy { // Inlined helper for pointer
							p := corev1.FSGroupChangeAlways
							return &p
						}(),
					},
					SetHostnameAsFQDN:             &setHostnameAsFQDN,
					TerminationGracePeriodSeconds: int64Ptr(terminationGracePeriodSeconds),
				},
			},
		},
	}

	// 4. Service
	serviceConstructID := appCfg.Service.Name + "-service"
	serviceName := generateCdk8sStyleName(chartID, serviceConstructID)
	serviceTargetPort := int32(appCfg.Service.HttpPort)
	serviceExposedPort := int32(80) // Renamed for clarity from servicePort

	service := &corev1.Service{
		TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "Service"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,
			Namespace: fullNsName,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{{
				Name:       "http", // Added port name
				Protocol:   corev1.ProtocolTCP, // Explicitly set
				Port:       serviceExposedPort,
				TargetPort: intstr.FromInt(int(serviceTargetPort)),
			}},
			Selector:    map[string]string{"cdk8s.io/metadata.addr": deploymentAddrLabelValue}, // Changed label key
			Type:        corev1.ServiceTypeClusterIP,
			// ExternalIPs: []string{}, // Omitting as it's default empty and kyaml might omit it. Test data has it.
		},
	}

	// 5. Ingress
	// Ingress name was "ingress" in cdk8s, not generated with suffix.
	// Host was "qwer." + k.host in cdk8s, but app.yaml has "simple-app.dev.tranquoctoan.com"
	// For now, using k.host directly as per problem description's app.yaml.
	// The cert-manager annotation also differs slightly.
	ingressName := "ingress"
	pathTypePrefix := networkingv1.PathTypePrefix
	// ingressHost variable is k.host. Test data expects "qwer." + k.host
	ingressRuleHost := "qwer." + k.host // To match testdata/app.yaml
	// TLS host in testdata is just k.host (treenq.com)
	ingressTLSHost := k.host


	ingress := &networkingv1.Ingress{
		TypeMeta: metav1.TypeMeta{APIVersion: "networking.k8s.io/v1", Kind: "Ingress"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      ingressName,
			Namespace: fullNsName,
			Annotations: map[string]string{
				"cert-manager.io/cluster-issuer": "letsencrypt-prod", // Changed to match testdata
			},
		},
		Spec: networkingv1.IngressSpec{
			Rules: []networkingv1.IngressRule{{
				Host: ingressRuleHost, // Changed to match testdata
				IngressRuleValue: networkingv1.IngressRuleValue{HTTP: &networkingv1.HTTPIngressRuleValue{
					Paths: []networkingv1.HTTPIngressPath{{
						Path:     "/",
						PathType: &pathTypePrefix,
						Backend: networkingv1.IngressBackend{
							Service: &networkingv1.IngressServiceBackend{
								Name: serviceName,
								Port: networkingv1.ServiceBackendPort{Number: serviceExposedPort}, // Use the service's exposed port
							},
						},
					},
				},
			},
		},
	}},
			TLS: []networkingv1.IngressTLS{{
				Hosts:      []string{ingressTLSHost}, // Changed to match testdata
				SecretName: "letsencrypt",
			}},
		},
	}
	// The original cdk8s code created two charts, one for ingress. Here we return all objects together.
	return []interface{}{namespace, registrySecret, deployment, service, ingress}, nil
}

// Helper functions for pointer types
func int32Ptr(i int32) *int32 { return &i }
func int64Ptr(i int64) *int64 { return &i }
func boolPtr(b bool) *bool    { return &b }


func (k *Kube) Apply(ctx context.Context, rawConig, data string) error {
	decoder := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme) // Stays yaml for this generic Apply func
	conf, err := clientcmd.RESTConfigFromKubeConfig([]byte(rawConig))
	if err != nil {
		return err
	}

	dynamicClient, err := dynamic.NewForConfig(conf)
	if err != nil {
		return fmt.Errorf("failed to create dynamic client: %w", err)
	}

	dataChunks := strings.Split(data, "---")
	var validObjs []*unstructured.Unstructured // Initialize empty slice

	for _, chunk := range dataChunks {
		trimmedChunk := strings.TrimSpace(chunk)
		if trimmedChunk == "" {
			continue
		}
		var obj unstructured.Unstructured
		_, _, err = decoder.Decode([]byte(trimmedChunk), nil, &obj)
		if err != nil {
			// If there's an error decoding (e.g. empty or malformed), skip this chunk.
			// This can happen with comments or empty lines between '---'
			fmt.Printf("Warning: skipping decoding of chunk due to error: %v\nChunk content:\n%s\n", err, trimmedChunk)
			continue
		}
		// Ensure that we only add non-empty objects.
		// Decoding an empty string might still produce an object with no GVK.
		if obj.GetKind() == "" && obj.GetAPIVersion() == "" {
			continue
		}
		validObjs = append(validObjs, &obj)
	}

	for _, obj := range validObjs {
		gvr, _ := meta.UnsafeGuessKindToResource(obj.GroupVersionKind())
		resourceClient := dynamicClient.Resource(gvr).Namespace(obj.GetNamespace())

		_, err = resourceClient.Create(ctx, obj, metav1.CreateOptions{})
		if errors.IsAlreadyExists(err) {
			// Attempt to get the existing object to retrieve its ResourceVersion for Update
			existingObj, getErr := resourceClient.Get(ctx, obj.GetName(), metav1.GetOptions{})
			if getErr != nil {
				return fmt.Errorf("failed to get existing object for update: %w", getErr)
			}
			obj.SetResourceVersion(existingObj.GetResourceVersion())

			_, err = resourceClient.Update(ctx, obj, metav1.UpdateOptions{})
			if err != nil {
				return fmt.Errorf("failed to update object: %w", err)
			}
		} else if err != nil {
			return fmt.Errorf("failed to create object: %w", err)
		}
	}

	return nil
}

func (k *Kube) StoreSecret(ctx context.Context, rawConfig string, space, repoID, key, value string) error {
	config, err := clientcmd.RESTConfigFromKubeConfig([]byte(rawConfig))
	if err != nil {
		return fmt.Errorf("failed to parse kubeconfig: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create kubernetes clientset: %w", err)
	}

	namespace := ns(space, repoID)
	secretObjectName := secretName(repoID, key)
	secretClient := clientset.CoreV1().Secrets(namespace)

	// Check if namespace exists, create if not
	_, err = clientset.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			_, createErr := clientset.CoreV1().Namespaces().Create(ctx, &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: namespace}}, metav1.CreateOptions{})
			if createErr != nil {
				return fmt.Errorf("failed to create namespace %s: %w", namespace, createErr)
			}
		} else {
			return fmt.Errorf("failed to get namespace %s: %w", namespace, err)
		}
	}

	existingSecret, err := secretClient.Get(ctx, secretObjectName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			newSecret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      secretObjectName,
					Namespace: namespace,
				},
				Data: map[string][]byte{
					key: []byte(value),
				},
			}
			_, err = secretClient.Create(ctx, newSecret, metav1.CreateOptions{})
			if err != nil {
				return fmt.Errorf("failed to create secret %s in namespace %s: %w", secretObjectName, namespace, err)
			}
			return nil
		}
		return fmt.Errorf("failed to get secret %s in namespace %s: %w", secretObjectName, namespace, err)
	}

	if existingSecret.Data == nil {
		existingSecret.Data = make(map[string][]byte)
	}
	existingSecret.Data[key] = []byte(value)

	_, err = secretClient.Update(ctx, existingSecret, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update secret %s in namespace %s: %w", secretObjectName, namespace, err)
	}

	return nil
}

func (k *Kube) GetSecret(ctx context.Context, rawConfig string, space, repoID, key string) (string, error) {
	config, err := clientcmd.RESTConfigFromKubeConfig([]byte(rawConfig))
	if err != nil {
		return "", fmt.Errorf("failed to parse kubeconfig: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return "", fmt.Errorf("failed to create kubernetes clientset: %w", err)
	}

	namespace := ns(space, repoID)
	secretObjectName := secretName(repoID, key)
	secretClient := clientset.CoreV1().Secrets(namespace)

	secret, err := secretClient.Get(ctx, secretObjectName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return "", domain.ErrSecretNotFound
		}
		return "", fmt.Errorf("failed to get secret %s in namespace %s: %w", secretObjectName, namespace, err)
	}
	if secret.Data == nil {
		secret.Data = make(map[string][]byte)
	}

	value, ok := secret.Data[key]
	if !ok {
		return "", domain.ErrSecretNotFound
	}

	return string(value), nil
}

func (k *Kube) RemoveSecret(ctx context.Context, rawConfig string, space, repoID, key string) error {
	config, err := clientcmd.RESTConfigFromKubeConfig([]byte(rawConfig))
	if err != nil {
		return fmt.Errorf("failed to parse kubeconfig: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create kubernetes clientset: %w", err)
	}

	namespace := ns(space, repoID)
	secretObjectName := secretName(repoID, key)
	secretClient := clientset.CoreV1().Secrets(namespace)

	err = secretClient.Delete(ctx, secretObjectName, metav1.DeleteOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to delete secret %s in namespace %s: %w", secretObjectName, namespace, err)
	}

	return nil
}
