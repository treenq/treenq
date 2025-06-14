package cdk

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	tqsdk "github.com/treenq/treenq/pkg/sdk"
	"github.com/treenq/treenq/src/domain"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
	kyaml "sigs.k8s.io/yaml"
)

const (
	defaultReplicas = 1
)

type Kube struct {
	host           string
	dockerRegistry string
	userName       string
	userPassword   string
}

func NewKube(
	host, dockerRegistry, userName, userPassword string,
) *Kube {
	return &Kube{host: host, dockerRegistry: dockerRegistry, userName: userName, userPassword: userPassword}
}

// DefineApp generates a Kubernetes manifest string for an application.
// It calls generateKubeResources to create Kubernetes objects and then serializes them to YAML.
// The ctx parameter is currently unused but kept for potential future use (e.g. logging, cancellation).
func (k *Kube) DefineApp(_ context.Context, id string, nsName string, app tqsdk.Space, image domain.Image, secretKeys []string) (string, error) {
	resources := k.generateKubeResources(id, nsName, app, image, secretKeys)

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
func (k *Kube) generateKubeResources(id, nsName string, app tqsdk.Space, image domain.Image, secretKeys []string) []any {
	fullNsName := ns(nsName, id)
	labels := map[string]string{"tq/name": app.Service.Name}

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
	registrySecret := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "Secret"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      registrySecretName,
			Namespace: fullNsName,
		},
		StringData: map[string]string{
			".dockerconfigjson": dockerConfigJSON,
		},
		Type: corev1.SecretTypeDockerConfigJson,
	}

	// 3. Deployment
	replicas := int32(defaultReplicas)
	if app.Service.Replicas > 0 {
		replicas = int32(app.Service.Replicas)
	}

	var envVars []corev1.EnvVar
	for key, value := range app.Service.RuntimeEnvs {
		envVars = append(envVars, corev1.EnvVar{Name: key, Value: value})
	}
	for _, key := range secretKeys {
		secretObjNameForRef := secretName(id, key)
		envVars = append(envVars, corev1.EnvVar{
			Name: strings.ToUpper(key),
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{Name: secretObjNameForRef},
					Key:                  key,
				},
			},
		})
	}

	computeRes := app.Service.ComputationResource
	readOnlyRootFilesystem := true
	runAsNonRoot := true
	runAsUser := int64(1000)

	cpuReqStr := "0m"
	memReqStr := "0Mi"
	ephemeralStorageReqStr := "0Gi"

	if computeRes.CpuUnits > 0 {
		cpuReqStr = fmt.Sprintf("%dm", computeRes.CpuUnits)
	}
	if computeRes.MemoryMibs > 0 {
		memReqStr = fmt.Sprintf("%dMi", computeRes.MemoryMibs)
	}
	if computeRes.DiskGibs > 0 {
		ephemeralStorageReqStr = fmt.Sprintf("%dGi", computeRes.DiskGibs)
	}

	deployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{APIVersion: "apps/v1", Kind: "Deployment"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      app.Service.Name,
			Namespace: fullNsName,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(replicas),
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:            app.Service.Name,
						Image:           image.FullPath(),
						ImagePullPolicy: corev1.PullAlways,
						Ports: []corev1.ContainerPort{{
							Name:          "http",
							ContainerPort: int32(app.Service.HttpPort),
						}},
						Env: envVars,
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:              resource.MustParse(cpuReqStr),
								corev1.ResourceMemory:           resource.MustParse(memReqStr),
								corev1.ResourceEphemeralStorage: resource.MustParse(ephemeralStorageReqStr),
							},
							Limits: corev1.ResourceList{
								corev1.ResourceCPU:              resource.MustParse(cpuReqStr),
								corev1.ResourceMemory:           resource.MustParse(memReqStr),
								corev1.ResourceEphemeralStorage: resource.MustParse(ephemeralStorageReqStr),
							},
						},
						SecurityContext: &corev1.SecurityContext{
							ReadOnlyRootFilesystem: &readOnlyRootFilesystem,
							RunAsNonRoot:           &runAsNonRoot,
							RunAsUser:              &runAsUser,
						},
					}},
					ImagePullSecrets: []corev1.LocalObjectReference{{Name: registrySecretName}},
					RestartPolicy:    corev1.RestartPolicyAlways,
					SecurityContext: &corev1.PodSecurityContext{
						RunAsUser:    &runAsUser,
						RunAsNonRoot: &runAsNonRoot,
						FSGroupChangePolicy: func() *corev1.PodFSGroupChangePolicy {
							p := corev1.FSGroupChangeAlways
							return &p
						}(),
					},
				},
			},
		},
	}

	// 4. Service
	serviceTargetPort := int32(app.Service.HttpPort)
	serviceExposedPort := int32(80)

	service := &corev1.Service{
		TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "Service"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      app.Service.Name,
			Namespace: fullNsName,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{{
				Name:       "http",
				Protocol:   corev1.ProtocolTCP,
				Port:       serviceExposedPort,
				TargetPort: intstr.FromInt(int(serviceTargetPort)),
			}},
			Selector: map[string]string{"tq/name": app.Service.Name},
			Type:     corev1.ServiceTypeClusterIP,
		},
	}

	// 5. Ingress
	ingressName := "ingress"
	pathTypePrefix := networkingv1.PathTypePrefix
	ingressRuleHost := "qwer." + k.host
	ingressTLSHost := k.host

	ingress := &networkingv1.Ingress{
		TypeMeta: metav1.TypeMeta{APIVersion: "networking.k8s.io/v1", Kind: "Ingress"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      ingressName,
			Namespace: fullNsName,
			Annotations: map[string]string{
				"cert-manager.io/cluster-issuer": "letsencrypt",
			},
		},
		Spec: networkingv1.IngressSpec{
			Rules: []networkingv1.IngressRule{{
				Host: ingressRuleHost,
				IngressRuleValue: networkingv1.IngressRuleValue{
					HTTP: &networkingv1.HTTPIngressRuleValue{
						Paths: []networkingv1.HTTPIngressPath{
							{
								Path:     "/",
								PathType: &pathTypePrefix,
								Backend: networkingv1.IngressBackend{
									Service: &networkingv1.IngressServiceBackend{
										Name: app.Service.Name,
										Port: networkingv1.ServiceBackendPort{Number: serviceExposedPort},
									},
								},
							},
						},
					},
				},
			}},
			TLS: []networkingv1.IngressTLS{{
				Hosts:      []string{ingressTLSHost},
				SecretName: "letsencrypt",
			}},
		},
	}
	return []any{namespace, registrySecret, deployment, service, ingress}
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
