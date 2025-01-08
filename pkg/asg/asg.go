package asg

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling/types"
	"github.com/guessi/eks-managed-node-groups/pkg/constants"
)

func GetAsgClient(region string) *autoscaling.Client {
	cfg, err := config.LoadDefaultConfig(
		context.Background(),
		config.WithRegion(region),
	)
	if err != nil {
		log.Fatalf("unable to load AWS SDK config, %v", err)
	}

	return autoscaling.NewFromConfig(cfg, func(options *autoscaling.Options) {
		options.APIOptions = append(
			options.APIOptions,
			middleware.AddUserAgentKeyValue(constants.AppName, constants.VersionString),
		)
	})
}

func GetAutoScalingGroupsByClusterName(client *autoscaling.Client, clusterName string) []string {
	result, err := client.DescribeAutoScalingGroups(context.Background(), &autoscaling.DescribeAutoScalingGroupsInput{
		Filters: []types.Filter{
			{
				Name:   aws.String(fmt.Sprintf("tag:kubernetes.io/cluster/%s", clusterName)),
				Values: []string{*aws.String("owned")},
			},
		}})
	if err != nil {
		log.Fatalf("unable to execute DescribeAutoScalingGroups, %v", err)
	}

	autoscalinggroups := []string{}
	for _, group := range result.AutoScalingGroups {
		var isManagedNodeGroup bool
		for _, tag := range group.Tags {
			if strings.Compare("eks:cluster-name", *aws.String(*tag.Key)) == 0 && strings.Compare(clusterName, *aws.String(*tag.Value)) == 0 {
				isManagedNodeGroup = true
				break
			}
		}
		if !isManagedNodeGroup {
			autoscalinggroups = append(autoscalinggroups, *group.AutoScalingGroupName)
		}
	}
	return autoscalinggroups
}

func DescribeAutoScalingGroupsByNodegroupName(client *autoscaling.Client, nodeGroupName string) (*autoscaling.DescribeAutoScalingGroupsOutput, error) {
	return client.DescribeAutoScalingGroups(context.Background(), &autoscaling.DescribeAutoScalingGroupsInput{AutoScalingGroupNames: []string{nodeGroupName}})
}

func SetDesiredCapacity(client *autoscaling.Client, setDesiredCapacityInput autoscaling.SetDesiredCapacityInput) (*autoscaling.SetDesiredCapacityOutput, error) {
	return client.SetDesiredCapacity(context.Background(), &setDesiredCapacityInput)
}
