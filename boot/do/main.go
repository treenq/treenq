package main

import (
	"github.com/pulumi/pulumi-digitalocean/sdk/v4/go/digitalocean"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		// VPC
		vpc, err := digitalocean.NewVpc(ctx, "tq-kube-vpc", &digitalocean.VpcArgs{
			Name:    pulumi.String("tq-kube-staging-vpc"),
			Region:  pulumi.String("fra1"),
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

		ctx.Export("kubeconfig", cluster.KubeConfigs.Index(pulumi.Int(0)).RawConfig())
		return nil
	})
}
