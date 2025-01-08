package eks

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/eks/types"
	"github.com/guessi/eks-managed-node-groups/pkg/constants"
)

func GetEksClient(region string) *eks.Client {
	cfg, err := config.LoadDefaultConfig(
		context.Background(),
		config.WithRegion(region),
	)
	if err != nil {
		log.Fatalf("unable to load AWS SDK config, %v", err)
	}

	return eks.NewFromConfig(cfg, func(options *eks.Options) {
		options.APIOptions = append(
			options.APIOptions,
			middleware.AddUserAgentKeyValue(constants.AppName, constants.VersionString),
		)
	})
}

func ListClusters(client *eks.Client) []string {
	result, err := client.ListClusters(context.Background(), &eks.ListClustersInput{})
	if err != nil {
		log.Fatalf("unable to execute ListClusters, %v", err)
	}
	return result.Clusters
}

func ListNodegroups(client *eks.Client, cluster string) []string {
	result, err := client.ListNodegroups(context.Background(), &eks.ListNodegroupsInput{ClusterName: &cluster})
	if err != nil {
		log.Fatalf("unable to execute ListNodegroups, %v", err)
	}
	return result.Nodegroups
}

func GetNodegroupScalingConfig(client *eks.Client, clusterName, nodegroupName string) (*types.NodegroupScalingConfig, error) {
	result, err := client.DescribeNodegroup(context.Background(), &eks.DescribeNodegroupInput{ClusterName: &clusterName, NodegroupName: &nodegroupName})
	if err != nil {
		return nil, err
	}
	return result.Nodegroup.ScalingConfig, nil
}

func UpdateNodegroupConfig(client *eks.Client, updateNodegroupConfigInput eks.UpdateNodegroupConfigInput) (*eks.UpdateNodegroupConfigOutput, error) {
	return client.UpdateNodegroupConfig(context.Background(), &updateNodegroupConfigInput)
}
