package cdk

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	cdk8splus "github.com/cdk8s-team/cdk8s-plus-go/cdk8splus31/v2"
	tqsdk "github.com/treenq/treenq/pkg/sdk"
	"github.com/treenq/treenq/src/domain"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
)

type Kube struct {
	// host holds a main app host, used to create sub hosts for a quick app preview
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

func (k *Kube) DefineApp(ctx context.Context, id string, app tqsdk.Space, image domain.Image) string {
	a := cdk8s.NewApp(nil)
	k.newAppChart(a, id, app, image)
	out := a.SynthYaml()
	return *out
}

func (k *Kube) newAppChart(scope constructs.Construct, id string, app tqsdk.Space, image domain.Image) []cdk8s.Chart {
	ns := jsii.String(app.Key + "-" + id)
	chart := cdk8s.NewChart(scope, jsii.String(app.Key), &cdk8s.ChartProps{
		Namespace: ns,
	})
	ingressChart := cdk8s.NewChart(scope, jsii.String(app.Key+"-ingress"), &cdk8s.ChartProps{})

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
		            }`, k.dockerRegistry, base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", k.userName, k.userPassword))))),
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
