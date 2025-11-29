package ui

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/eks/types"
	"github.com/charmbracelet/huh"
	asgwrapper "github.com/guessi/eks-managed-node-groups/pkg/asg"
	"github.com/guessi/eks-managed-node-groups/pkg/constants"
	ekswrapper "github.com/guessi/eks-managed-node-groups/pkg/eks"
	"github.com/guessi/eks-managed-node-groups/pkg/utils"
)

func ShowVersion() {
	r, _ := regexp.Compile(`v[0-9]\.[0-9]+\.[0-9]+`)
	versionInfo := r.FindString(constants.GitVersion)
	fmt.Println(constants.AppName, versionInfo)
	fmt.Println(" Git Commit:", constants.GitVersion)
	fmt.Println(" Build with:", constants.GoVersion)
	fmt.Println(" Build time:", constants.BuildTime)
}

func clustersForm(clusters []string) (string, error) {
	var clusterName string

	clusterForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title(fmt.Sprintf("Choose your cluster (Total: %d)", len(clusters))).
				Options(huh.NewOptions(clusters...)...).
				Value(&clusterName).
				Height(10),
		),
	)
	err := clusterForm.Run()
	if err != nil {
		return "", fmt.Errorf("cluster form error: %w", err)
	}

	return clusterName, nil
}

func nodeGroupTypeForm() (string, error) {
	var targetType string
	targetTypeForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("What kind of node group it is about?").
				Options(
					huh.NewOption(
						constants.NodeGroupTypes[constants.Managed],
						constants.NodeGroupTypes[constants.Managed],
					),
					huh.NewOption(
						constants.NodeGroupTypes[constants.SelfManaged],
						constants.NodeGroupTypes[constants.SelfManaged],
					),
				).
				Value(&targetType).
				Height(10),
		),
	)
	err := targetTypeForm.Run()
	if err != nil {
		return "", fmt.Errorf("node group type form error: %w", err)
	}

	return targetType, nil
}

func nodegroupsForm(nodegroups []string) (string, error) {
	var nodegroupName string

	nodegroupForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title(fmt.Sprintf("Choose your nodegroups (Total: %d)", len(nodegroups))).
				Options(huh.NewOptions(nodegroups...)...).
				Value(&nodegroupName).
				Height(10),
		),
	)
	err := nodegroupForm.Run()
	if err != nil {
		return "", fmt.Errorf("nodegroup form error: %w", err)
	}

	return nodegroupName, nil
}

func nodegroupSizeForm() (int32, int32, int32, error) {
	var desiredSize, minSize, maxSize string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Desired size").
				Description("Desired size of node group?").
				Value(&desiredSize).
				Validate(utils.IsInteger).
				CharLimit(3),
			huh.NewInput().
				Title("Min size").
				Description("Min size of node group?").
				Value(&minSize).
				Validate(utils.IsInteger).
				CharLimit(3),
			huh.NewInput().
				Title("Max size").
				Description("Max size of node group?").
				Value(&maxSize).
				Validate(utils.IsInteger).
				CharLimit(3),
		),
	)
	if err := form.Run(); err != nil {
		return 0, 0, 0, err
	}

	desired, _ := utils.ParseInt32(desiredSize)
	min, _ := utils.ParseInt32(minSize)
	max, _ := utils.ParseInt32(maxSize)

	return desired, min, max, nil
}

func selfManagedNodeGroupWorkflow(asgClient *autoscaling.Client, clusterName string) error {
	nodegroups, err := asgwrapper.GetAutoScalingGroupsByClusterName(asgClient, clusterName)
	if err != nil {
		return err
	}
	if len(nodegroups) == 0 {
		fmt.Println("no nodegroup found")
		return nil
	}
	nodegroupName, err := nodegroupsForm(nodegroups)
	if err != nil {
		return err
	}

	desiredSize, minSize, maxSize, err := nodegroupSizeForm()
	if err != nil {
		return err
	}
	if err := utils.ValidateNodegroupSize(desiredSize, minSize, maxSize); err != nil {
		return err
	}

	describeAutoScalingGroupsOutput, err := asgwrapper.DescribeAutoScalingGroupsByNodegroupName(asgClient, nodegroupName)
	if err != nil {
		return err
	}

	var currentDesiredCapacity, currentMinSize, currentMaxSize int32
	for _, group := range describeAutoScalingGroupsOutput.AutoScalingGroups {
		if strings.Compare(*group.AutoScalingGroupName, nodegroupName) == 0 {
			currentDesiredCapacity = *group.DesiredCapacity
			currentMinSize = *group.MinSize
			currentMaxSize = *group.MaxSize
		}
	}

	if currentDesiredCapacity == desiredSize && currentMinSize == minSize && currentMaxSize == maxSize {
		fmt.Println("no change required, target node group size have no difference")
		return nil
	}

	updateAutoScalingGroupInput := autoscaling.UpdateAutoScalingGroupInput{
		AutoScalingGroupName: &nodegroupName,
		DesiredCapacity:      &desiredSize,
		MinSize:              &minSize,
		MaxSize:              &maxSize,
	}

	_, err = asgwrapper.UpdateAutoScalingGroup(asgClient, updateAutoScalingGroupInput)
	if err != nil {
		return err
	}

	fmt.Printf("Request details: {clusterName: %s, nodegroupName: %s, desiredSize: %d, minSize: %d, maxSize: %d}\n", clusterName, nodegroupName, desiredSize, minSize, maxSize)
	fmt.Printf("Request sent at: %s\n", time.Now().Format(time.RFC3339))

	return nil
}

func managedNodeGroupWorkflow(eksClient *eks.Client, clusterName string) error {
	nodegroups, err := ekswrapper.ListNodegroups(eksClient, clusterName)
	if err != nil {
		return err
	}
	if len(nodegroups) == 0 {
		fmt.Println("no nodegroup found")
		return nil
	}
	nodegroupName, err := nodegroupsForm(nodegroups)
	if err != nil {
		return err
	}

	desiredSize, minSize, maxSize, err := nodegroupSizeForm()
	if err != nil {
		return err
	}
	if err := utils.ValidateNodegroupSize(desiredSize, minSize, maxSize); err != nil {
		return err
	}

	scalingConfig, err := ekswrapper.GetNodegroupScalingConfig(eksClient, clusterName, nodegroupName)
	if err != nil {
		return err
	}
	if *scalingConfig.DesiredSize == desiredSize && *scalingConfig.MinSize == minSize && *scalingConfig.MaxSize == maxSize {
		fmt.Println("no change required, target node group size have no difference")
		return nil
	}

	updateNodegroupConfigInput := eks.UpdateNodegroupConfigInput{
		ClusterName:   &clusterName,
		NodegroupName: &nodegroupName,
		ScalingConfig: &types.NodegroupScalingConfig{DesiredSize: &desiredSize, MinSize: &minSize, MaxSize: &maxSize},
	}

	result, err := ekswrapper.UpdateNodegroupConfig(eksClient, updateNodegroupConfigInput)
	if err != nil {
		return err
	}

	fmt.Printf("Request details: {clusterName: %s, nodegroupName: %s, desiredSize: %d, minSize: %d, maxSize: %d}\n", clusterName, nodegroupName, desiredSize, minSize, maxSize)
	fmt.Printf("Request sent at: %s\n", result.Update.CreatedAt.Format(time.RFC3339))

	return nil
}

func Entry(region string) error {
	eksClient, err := ekswrapper.GetEksClient(region)
	if err != nil {
		return err
	}

	clusters, err := ekswrapper.ListClusters(eksClient)
	if err != nil {
		return err
	}
	if len(clusters) == 0 {
		fmt.Println("no cluster found")
		return nil
	}
	clusterName, err := clustersForm(clusters)
	if err != nil {
		return err
	}

	nodeGroupType, err := nodeGroupTypeForm()
	if err != nil {
		return err
	}

	if nodeGroupType == constants.NodeGroupTypes[constants.SelfManaged] {
		asgClient, err := asgwrapper.GetAsgClient(region)
		if err != nil {
			return err
		}

		if err := selfManagedNodeGroupWorkflow(asgClient, clusterName); err != nil {
			return err
		}
	} else {
		if err := managedNodeGroupWorkflow(eksClient, clusterName); err != nil {
			return err
		}
	}
	return nil
}
