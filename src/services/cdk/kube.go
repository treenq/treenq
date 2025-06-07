package cdk

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"sync"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	cdk8splus "github.com/cdk8s-team/cdk8s-plus-go/cdk8splus31/v2"
	tqsdk "github.com/treenq/treenq/pkg/sdk"
	"github.com/treenq/treenq/src/domain"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type Kube struct {
	// host holds a main app host, used to create sub hosts for a quick app preview
	host           string
	dockerRegistry string
	userName       string
	userPassword   string

	// jsii is not thread safe
	mx sync.Mutex
}

func NewKube(
	host, dockerRegistry, userName, userPassword string,
) *Kube {
	return &Kube{host: host, dockerRegistry: dockerRegistry, userName: userName, userPassword: userPassword}
}

func (k *Kube) DefineApp(ctx context.Context, id string, nsName string, app tqsdk.Space, image domain.Image, secretKeys []string) string {
	k.mx.Lock()
	defer k.mx.Unlock()
	a := cdk8s.NewApp(nil)
	k.newAppChart(a, id, nsName, app, image, secretKeys)
	out := a.SynthYaml()
	return *out
}

func ns(space, repoID string) string {
	return space + "-" + repoID
}

func secretName(repoID, key string) string {
	return repoID + "-" + strings.ToLower(key)
}

func (k *Kube) newAppChart(scope constructs.Construct, id, nsName string, app tqsdk.Space, image domain.Image, secretKeys []string) []cdk8s.Chart {
	ns := jsii.String(ns(nsName, id))
	chart := cdk8s.NewChart(scope, jsii.String(nsName), &cdk8s.ChartProps{
		Namespace: ns,
	})
	ingressChart := cdk8s.NewChart(scope, jsii.String(nsName+"-ingress"), &cdk8s.ChartProps{})

	cdk8splus.NewNamespace(chart, jsii.String(id+"-ns"), &cdk8splus.NamespaceProps{
		Metadata: &cdk8s.ApiObjectMetadata{
			Name:      ns,
			Namespace: jsii.String(""),
		},
	})

	envs := make(map[string]cdk8splus.EnvValue)
	for k, v := range app.Service.RuntimeEnvs {
		envs[k] = cdk8splus.EnvValue_FromValue(jsii.String(v))
	}
	for i := range secretKeys {
		secretObjectName := secretName(id, secretKeys[i])
		envs[secretKeys[i]] = cdk8splus.EnvValue_FromSecretValue(&cdk8splus.SecretValue{
			Key:    jsii.String(secretKeys[i]),
			Secret: cdk8splus.Secret_FromSecretName(scope, jsii.String(id), &secretObjectName),
		}, &cdk8splus.EnvValueFromSecretOptions{})
	}

	computeRes := app.Service.ComputationResource

	// tmpVolume := cdk8splus.Volume_FromEmptyDir(chart, jsii.String(app.Service.Name+"-volume-tmp"), jsii.String("tmp"), nil)
	// nginxVolume := cdk8splus.Volume_FromEmptyDir(chart, jsii.String(app.Service.Name+"-nginx"), jsii.String("nginx"), nil)

	registrySecret := cdk8splus.NewSecret(chart, jsii.String("registry-secret"), &cdk8splus.SecretProps{
		Metadata: &cdk8s.ApiObjectMetadata{
			Name: jsii.String("registry-credentials"),
		},
		StringData: &map[string]*string{
			".dockerconfigjson": jsii.String(fmt.Sprintf(`{
		                "auths": {
		                    "%s": {
		                        "auth": "%s"
		                    }
		                }
		            }`, k.dockerRegistry, base64.StdEncoding.EncodeToString(fmt.Appendf(nil, "%s:%s", k.userName, k.userPassword)))),
		},
		Type: jsii.String("kubernetes.io/dockerconfigjson"),
	})

	deployment := cdk8splus.NewDeployment(chart, jsii.String(app.Service.Name+"-deployment"), &cdk8splus.DeploymentProps{
		Replicas: jsii.Number(app.Service.Replicas),
		Containers: &[]*cdk8splus.ContainerProps{{
			Name:  jsii.String(app.Service.Name),
			Image: jsii.String(image.FullPath()),
			Ports: &[]*cdk8splus.ContainerPort{{
				Number: jsii.Number(app.Service.HttpPort),
				Name:   jsii.String("http"),
			}},
			EnvVariables: &envs,
			// VolumeMounts: &[]*cdk8splus.VolumeMount{
			// 	{
			// 		Path:   jsii.String("/tmp"),
			// 		Volume: tmpVolume,
			// 	},
			// 	{
			// 		Path:   jsii.String("/var"),
			// 		Volume: nginxVolume,
			// 	},
			// },
			Resources: &cdk8splus.ContainerResources{
				Cpu: &cdk8splus.CpuResources{
					Limit:   cdk8splus.Cpu_Millis(jsii.Number(computeRes.CpuUnits)),
					Request: cdk8splus.Cpu_Millis(jsii.Number(computeRes.CpuUnits)),
				},
				EphemeralStorage: &cdk8splus.EphemeralStorageResources{
					Limit:   cdk8s.Size_Gibibytes(jsii.Number(computeRes.DiskGibs)),
					Request: cdk8s.Size_Gibibytes(jsii.Number(computeRes.DiskGibs)),
				},
				Memory: &cdk8splus.MemoryResources{
					Limit:   cdk8s.Size_Mebibytes(jsii.Number(computeRes.MemoryMibs)),
					Request: cdk8s.Size_Mebibytes(jsii.Number(computeRes.MemoryMibs)),
				},
			},
			SecurityContext: &cdk8splus.ContainerSecurityContextProps{
				AllowPrivilegeEscalation: jsii.Bool(false),
				EnsureNonRoot:            jsii.Bool(true),
				Privileged:               jsii.Bool(false),
				ReadOnlyRootFilesystem:   jsii.Bool(true),
				User:                     jsii.Number(1000),
			},
		}},
		// Volumes: &[]cdk8splus.Volume{tmpVolume, nginxVolume},
		SecurityContext: &cdk8splus.PodSecurityContextProps{
			EnsureNonRoot:       jsii.Bool(true),
			FsGroupChangePolicy: cdk8splus.FsGroupChangePolicy_ALWAYS,
			User:                jsii.Number(1000),
		},
	})

	deployment.ApiObject().AddJsonPatch(cdk8s.JsonPatch_Add(
		jsii.String("/spec/template/spec/imagePullSecrets"),
		[]map[string]string{{"name": *registrySecret.Name()}},
	))

	servicePort := jsii.Number(80)

	service := cdk8splus.NewService(chart, jsii.String(app.Service.Name+"-service"), &cdk8splus.ServiceProps{
		Ports: &[]*cdk8splus.ServicePort{{
			Name:       jsii.String("http"),
			Port:       servicePort,
			TargetPort: jsii.Number(app.Service.HttpPort),
		}},
		Selector: deployment,
	})

	// define 3d level domain given from the existing domain
	cdk8splus.NewIngress(chart, jsii.String("ingress"), &cdk8splus.IngressProps{
		Rules: &[]*cdk8splus.IngressRule{{
			Host:     jsii.String("qwer" + "." + k.host),
			Path:     jsii.String("/"),
			PathType: cdk8splus.HttpIngressPathType_PREFIX,
			Backend: cdk8splus.IngressBackend_FromService(service, &cdk8splus.ServiceIngressBackendOptions{
				Port: servicePort,
			}),
		}},
	})

	return []cdk8s.Chart{chart, ingressChart}
}

func (k *Kube) Apply(ctx context.Context, rawConig, data string) error {
	decoder := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	conf, err := clientcmd.RESTConfigFromKubeConfig([]byte(rawConig))
	if err != nil {
		return err
	}

	dynamicClient, err := dynamic.NewForConfig(conf)
	if err != nil {
		return fmt.Errorf("failed to create dynamic client: %w", err)
	}

	dataChunks := strings.Split(data, "---")

	objs := make([]*unstructured.Unstructured, len(dataChunks))
	for i, chunk := range dataChunks {
		var obj unstructured.Unstructured
		_, _, err = decoder.Decode([]byte(chunk), nil, &obj)
		if err != nil {
			return fmt.Errorf("failed to decode YAML: %w", err)
		}
		objs[i] = &obj
	}
	for _, obj := range objs {
		gvr, _ := meta.UnsafeGuessKindToResource(obj.GroupVersionKind())
		resourceClient := dynamicClient.Resource(gvr).Namespace(obj.GetNamespace())

		_, err = resourceClient.Create(ctx, obj, metav1.CreateOptions{})
		if errors.IsAlreadyExists(err) {
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
