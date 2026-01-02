package eks

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/eks/types"
	"github.com/guessi/eks-managed-node-groups/pkg/constants"
)

func ValidateCredentials(region string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return fmt.Errorf("unable to load AWS config: %w", err)
	}

	return nil
}

func GetEksClient(region string) (*eks.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion(region),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to load AWS SDK config: %w", err)
	}

	return eks.NewFromConfig(cfg, func(options *eks.Options) {
		options.APIOptions = append(
			options.APIOptions,
			middleware.AddUserAgentKeyValue(constants.AppName, constants.GitVersion),
		)
	}), nil
}

func ListClusters(client *eks.Client) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := client.ListClusters(ctx, &eks.ListClustersInput{})
	if err != nil {
		return nil, fmt.Errorf("unable to execute ListClusters: %w", err)
	}
	return result.Clusters, nil
}

func ListNodegroups(client *eks.Client, cluster string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := client.ListNodegroups(ctx, &eks.ListNodegroupsInput{ClusterName: &cluster})
	if err != nil {
		return nil, fmt.Errorf("unable to execute ListNodegroups: %w", err)
	}
	return result.Nodegroups, nil
}

func GetNodegroupScalingConfig(client *eks.Client, clusterName, nodegroupName string) (*types.NodegroupScalingConfig, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := client.DescribeNodegroup(ctx, &eks.DescribeNodegroupInput{ClusterName: &clusterName, NodegroupName: &nodegroupName})
	if err != nil {
		return nil, err
	}
	return result.Nodegroup.ScalingConfig, nil
}

func UpdateNodegroupConfig(client *eks.Client, updateNodegroupConfigInput eks.UpdateNodegroupConfigInput) (*eks.UpdateNodegroupConfigOutput, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return client.UpdateNodegroupConfig(ctx, &updateNodegroupConfigInput)
}
