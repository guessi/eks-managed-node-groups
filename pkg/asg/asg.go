package asg

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling/types"
	"github.com/guessi/eks-managed-node-groups/pkg/constants"
)

func GetAsgClient(region string) (*autoscaling.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion(region),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to load AWS SDK config: %w", err)
	}

	return autoscaling.NewFromConfig(cfg, func(options *autoscaling.Options) {
		options.APIOptions = append(
			options.APIOptions,
			middleware.AddUserAgentKeyValue(constants.AppName, constants.GitVersion),
		)
	}), nil
}

func isManagedNodeGroup(group types.AutoScalingGroup, clusterName string) bool {
	hasClusterTag := false
	hasNodegroupTag := false

	for _, tag := range group.Tags {
		if tag.Key != nil && tag.Value != nil {
			if *tag.Key == "eks:cluster-name" && *tag.Value == clusterName {
				hasClusterTag = true
			}
			if *tag.Key == "eks:nodegroup-name" {
				hasNodegroupTag = true
			}
		}
	}

	return hasClusterTag && hasNodegroupTag
}

func GetAutoScalingGroupsByClusterName(client *autoscaling.Client, clusterName string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := client.DescribeAutoScalingGroups(ctx, &autoscaling.DescribeAutoScalingGroupsInput{
		Filters: []types.Filter{
			{
				Name:   aws.String(fmt.Sprintf("tag:kubernetes.io/cluster/%s", clusterName)),
				Values: []string{"owned"},
			},
		}})
	if err != nil {
		return nil, fmt.Errorf("unable to execute DescribeAutoScalingGroups: %w", err)
	}

	autoscalinggroups := []string{}
	for _, group := range result.AutoScalingGroups {
		if group.AutoScalingGroupName == nil {
			continue
		}
		if !isManagedNodeGroup(group, clusterName) {
			autoscalinggroups = append(autoscalinggroups, *group.AutoScalingGroupName)
		}
	}
	return autoscalinggroups, nil
}

func DescribeAutoScalingGroupsByNodegroupName(client *autoscaling.Client, nodeGroupName string) (*autoscaling.DescribeAutoScalingGroupsOutput, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return client.DescribeAutoScalingGroups(ctx, &autoscaling.DescribeAutoScalingGroupsInput{AutoScalingGroupNames: []string{nodeGroupName}})
}

func UpdateAutoScalingGroup(client *autoscaling.Client, updateAutoScalingGroupInput autoscaling.UpdateAutoScalingGroupInput) (*autoscaling.UpdateAutoScalingGroupOutput, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return client.UpdateAutoScalingGroup(ctx, &updateAutoScalingGroupInput)
}
