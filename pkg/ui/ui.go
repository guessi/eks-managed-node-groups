package ui

import (
	"fmt"
	"log"
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

func clustersForm(clusters []string) string {
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
		log.Fatal(err)
	}

	return clusterName
}

func nodeGroupTypeForm() string {
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
		log.Fatal(err)
	}

	return targetType
}

func nodegroupsForm(nodegroups []string) string {
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
		log.Fatal(err)
	}

	return nodegroupName
}

func nodegroupSizeForm() (int32, int32, int32) {
	var desiredSize, minSize, maxSize string

	nodegroupSizeForm := huh.NewForm(
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
	err := nodegroupSizeForm.Run()
	if err != nil {
		log.Fatal(err)
	}

	return utils.ParseInt32(desiredSize), utils.ParseInt32(minSize), utils.ParseInt32(maxSize)
}

func nodegroupDesiredCapacityForm() int32 {
	var desiredCapacity string

	nodegroupDesiredCapacityForm := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Desired capacity").
				Description("Desired capacity of node group?").
				Value(&desiredCapacity).
				Validate(utils.IsInteger).
				CharLimit(3),
		),
	)
	err := nodegroupDesiredCapacityForm.Run()
	if err != nil {
		log.Fatal(err)
	}

	return utils.ParseInt32(desiredCapacity)
}

func selfManagedNodeGroupWorkflow(asgClient *autoscaling.Client, clusterName string) error {
	nodegroups := asgwrapper.GetAutoScalingGroupsByClusterName(asgClient, clusterName)
	if len(nodegroups) == 0 {
		fmt.Println("no nodegroup found")
		return nil
	}
	nodegroupName := nodegroupsForm(nodegroups)

	desiredCapacity := nodegroupDesiredCapacityForm()

	if err := utils.ValidateDesiredCapacity(desiredCapacity); err != nil {
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

	if desiredCapacity < currentMinSize {
		return fmt.Errorf("desired capacity is lower than target node group's min size, current state: {desired: %d, min: %d, max: %d}", currentDesiredCapacity, currentMinSize, currentMaxSize)
	}

	if desiredCapacity < currentMaxSize {
		return fmt.Errorf("desired capacity is larger than target node group's max size, current state: {desired: %d, min: %d, max: %d}", currentDesiredCapacity, currentMinSize, currentMaxSize)
	}

	if currentDesiredCapacity == desiredCapacity {
		fmt.Println("no change required, target node group size have no difference")
		return nil
	}

	setDesiredCapacityInput := autoscaling.SetDesiredCapacityInput{
		AutoScalingGroupName: &nodegroupName,
		DesiredCapacity:      &desiredCapacity,
	}

	_, err = asgwrapper.SetDesiredCapacity(asgClient, setDesiredCapacityInput)
	if err != nil {
		return err
	}

	fmt.Printf("Request details: {desiredCapacity: %d}\n", desiredCapacity)

	return nil
}

func managedNodeGroupWorkflow(eksClient *eks.Client, clusterName string) error {
	nodegroups := ekswrapper.ListNodegroups(eksClient, clusterName)
	if len(nodegroups) == 0 {
		fmt.Println("no nodegroup found")
		return nil
	}
	nodegroupName := nodegroupsForm(nodegroups)

	desiredSize, minSize, maxSize := nodegroupSizeForm()
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
	eksClient := ekswrapper.GetEksClient(region)

	clusters := ekswrapper.ListClusters(eksClient)
	if len(clusters) == 0 {
		fmt.Println("no cluster found")
		return nil
	}
	clusterName := clustersForm(clusters)

	nodeGroupType := nodeGroupTypeForm()

	if nodeGroupType == constants.NodeGroupTypes[constants.SelfManaged] {
		asgClient := asgwrapper.GetAsgClient(region)

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
