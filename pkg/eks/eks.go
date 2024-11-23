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

func GetEksClient() *eks.Client {
	cfg, err := config.LoadDefaultConfig(
		context.Background(),
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

func GetNodegroupScalingConfig(client *eks.Client, clusterName, nodegroupName string) (int32, int32, int32) {
	result, err := client.DescribeNodegroup(context.Background(), &eks.DescribeNodegroupInput{ClusterName: &clusterName, NodegroupName: &nodegroupName})
	if err != nil {
		log.Fatal(err)
	}
	return *result.Nodegroup.ScalingConfig.DesiredSize, *result.Nodegroup.ScalingConfig.MinSize, *result.Nodegroup.ScalingConfig.MaxSize
}

func UpdateNodegroupConfig(client *eks.Client, clusterName, nodegroupName string, desired, min, max int32) (*eks.UpdateNodegroupConfigOutput, error) {
	scalingConfig := types.NodegroupScalingConfig{
		DesiredSize: &desired,
		MinSize:     &min,
		MaxSize:     &max,
	}
	updateNodegroupConfigInput := eks.UpdateNodegroupConfigInput{
		ClusterName:   &clusterName,
		NodegroupName: &nodegroupName,
		ScalingConfig: &scalingConfig,
	}
	return client.UpdateNodegroupConfig(context.Background(), &updateNodegroupConfigInput)
}
