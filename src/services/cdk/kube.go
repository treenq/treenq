package cdk

import (
	"context"
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
}

func NewKube() *Kube {
	return &Kube{}
}

func (k *Kube) DefineApp(ctx context.Context, id string, app tqsdk.Space, image domain.Image) string {
	a := cdk8s.NewApp(nil)
	k.newAppChart(a, id, app, image)
	out := a.SynthYaml()
	return *out
}

func (k *Kube) newAppChart(scope constructs.Construct, id string, app tqsdk.Space, image domain.Image) cdk8s.Chart {
	ns := jsii.String(id + "-" + app.Key)
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		Namespace: ns,
	})

	cdk8splus.NewNamespace(chart, jsii.String(id), &cdk8splus.NamespaceProps{
		Metadata: &cdk8s.ApiObjectMetadata{
			Name:      ns,
			Namespace: jsii.String(""),
		},
	})

	envs := make(map[string]cdk8splus.EnvValue)
	for k, v := range app.Service.RuntimeEnvs {
		envs[k] = cdk8splus.EnvValue_FromValue(jsii.String(v))
	}
	computeRes := app.Service.SizeSlug.ToComputationResource()
	deployment := cdk8splus.NewDeployment(chart, jsii.String("deployment"), &cdk8splus.DeploymentProps{
		Replicas: jsii.Number(app.Service.Repicas),
		Containers: &[]*cdk8splus.ContainerProps{{
			Name:  jsii.String(app.Service.Name),
			Image: jsii.String(image.FullPath()),
			Ports: &[]*cdk8splus.ContainerPort{{
				Number: jsii.Number(app.Service.HttpPort),
				Name:   jsii.String("http"),
			}},
			EnvVariables: &envs,
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
		}},
	})

	service := cdk8splus.NewService(chart, jsii.String("service"), &cdk8splus.ServiceProps{
		Ports: &[]*cdk8splus.ServicePort{{
			Name:       jsii.String("http"),
			Port:       jsii.Number(80),
			TargetPort: jsii.Number(app.Service.HttpPort),
		}},
		Selector: deployment,
	})

	cdk8splus.NewIngress(chart, jsii.String("ingress"), &cdk8splus.IngressProps{
		// Metadata: &cdk8s.ApiObjectMetadata{
		// 	Annotations: &map[string]*string{
		// 		"nginx.ingress.kubernetes.io/rewrite-target": jsii.String("/"),
		// 	},
		// },
		Rules: &[]*cdk8splus.IngressRule{{
			Path:     jsii.String("/"),
			PathType: cdk8splus.HttpIngressPathType_PREFIX,
			Backend:  cdk8splus.IngressBackend_FromResource(service),
		}},
	})

	return chart
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
