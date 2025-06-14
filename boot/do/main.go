package main

import (
	"fmt"

	"github.com/pulumi/pulumi-digitalocean/sdk/v4/go/digitalocean"
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/yaml"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		// VPC
		vpc, err := digitalocean.NewVpc(ctx, "tq-kube-vpc", &digitalocean.VpcArgs{
			Name:    pulumi.String("tq-kube-staging-vpc"),
			Region:  digitalocean.RegionFRA1,
			IpRange: pulumi.String("10.10.0.0/16"),
		})
		if err != nil {
			return err
		}

		// Kubernetes Cluster
		cluster, err := digitalocean.NewKubernetesCluster(ctx, "my-cluster", &digitalocean.KubernetesClusterArgs{
			Region:  digitalocean.RegionFRA1,
			VpcUuid: vpc.ID(),
			Version: pulumi.String("1.32.2-do.3"),
			NodePool: &digitalocean.KubernetesClusterNodePoolArgs{
				Name:      pulumi.String("tq-staging-node-pool"),
				Size:      digitalocean.DropletSlugDropletS1VCPU2GB,
				NodeCount: pulumi.Int(1),
			},
		})
		if err != nil {
			return err
		}
		kubeconfig := cluster.KubeConfigs.Index(pulumi.Int(0)).RawConfig()
		prov, _ := kubernetes.NewProvider(ctx, "k8s", &kubernetes.ProviderArgs{Kubeconfig: kubeconfig})

		_, err = helm.NewRelease(ctx, "ingress-nginx", &helm.ReleaseArgs{
			Chart: pulumi.String("ingress-nginx"),
			RepositoryOpts: &helm.RepositoryOptsArgs{
				Repo: pulumi.String("https://kubernetes.github.io/ingress-nginx"),
			},
			Namespace:       pulumi.String("ingress-nginx"),
			CreateNamespace: pulumi.Bool(true),
		}, pulumi.Provider(prov))
		if err != nil {
			return fmt.Errorf("failed to define ingress release: %w", err)
		}

		_, err = helm.NewRelease(ctx, "cert-manager", &helm.ReleaseArgs{
			Chart: pulumi.String("cert-manager"),
			RepositoryOpts: &helm.RepositoryOptsArgs{
				Repo: pulumi.String("https://charts.jetstack.io"),
			},
			Namespace: pulumi.String("cert-manager"),
			Values: pulumi.Map{
				"installCRDs": pulumi.BoolPtr(true),
			},
			CreateNamespace: pulumi.Bool(true),
		}, pulumi.Provider(prov))
		if err != nil {
			return fmt.Errorf("failed to define cert manager: %w", err)
		}

		_, err = yaml.NewConfigFile(ctx, "letsencrypt-issuer", &yaml.ConfigFileArgs{
			File: "issuer.yaml",
		}, pulumi.Provider(prov))
		if err != nil {
			return fmt.Errorf("failed to define letsencrypt issuer: %w", err)
		}

		containerRegistry, err := digitalocean.NewContainerRegistry(ctx, "tq-staging", &digitalocean.ContainerRegistryArgs{
			Name:                 pulumi.String("tq-staging"),
			SubscriptionTierSlug: pulumi.String("starter"),
			Region:               digitalocean.RegionFRA1,
		})
		if err != nil {
			return fmt.Errorf("failed to define container registry: %w", err)
		}

		ctx.Export("kubeconfig", kubeconfig)
		ctx.Export("registry_url", containerRegistry.Endpoint)
		return nil
	})
}
