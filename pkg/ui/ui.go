package ui

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"slices"
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

type Options struct {
	Region        string
	Profile       string
	ClusterName   string
	NodeGroupType string // "", "managed" or "self-managed"
	NodegroupName string
	DesiredSize   *int32
	MinSize       *int32
	MaxSize       *int32
	DryRun        bool
	Yes           bool
	Output        string // "text" or "json"
}

var regionPattern = regexp.MustCompile(`^[a-z0-9-]+$`)

func (o Options) Validate() error {
	if o.Region != "" && !regionPattern.MatchString(o.Region) {
		return fmt.Errorf("invalid --region %q, must match %q", o.Region, regionPattern.String())
	}

	if o.NodeGroupType != "" && o.NodeGroupType != "managed" && o.NodeGroupType != "self-managed" {
		return fmt.Errorf("invalid --type %q, must be one of: managed, self-managed", o.NodeGroupType)
	}

	if o.Output != "" && o.Output != "text" && o.Output != "json" {
		return fmt.Errorf("invalid --output %q, must be one of: text, json", o.Output)
	}

	sizesSet := 0
	for _, size := range []*int32{o.DesiredSize, o.MinSize, o.MaxSize} {
		if size != nil {
			sizesSet++
		}
	}
	if sizesSet != 0 && sizesSet != 3 {
		return errors.New("--desired, --min and --max must be provided together")
	}

	return nil
}

type changeReport struct {
	Cluster        string `json:"cluster"`
	Nodegroup      string `json:"nodegroup"`
	CurrentDesired int32  `json:"currentDesiredSize"`
	CurrentMin     int32  `json:"currentMinSize"`
	CurrentMax     int32  `json:"currentMaxSize"`
	DesiredSize    int32  `json:"desiredSize"`
	MinSize        int32  `json:"minSize"`
	MaxSize        int32  `json:"maxSize"`
	Applied        bool   `json:"applied"`
	SentAt         string `json:"sentAt,omitempty"`
}

func printJSON(v any) error {
	out, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("unable to marshal output: %w", err)
	}
	fmt.Println(string(out))
	return nil
}

func printStatus(opts Options, message string) error {
	if opts.Output == "json" {
		return printJSON(struct {
			Message string `json:"message"`
			Applied bool   `json:"applied"`
		}{message, false})
	}
	fmt.Println(message)
	return nil
}

// scalingConfigDrifted reports whether the latest size values no longer match
// the baseline values shown to the user before confirmation
func scalingConfigDrifted(baseDesired, baseMin, baseMax int32, latestDesired, latestMin, latestMax *int32) bool {
	return latestDesired == nil || latestMin == nil || latestMax == nil ||
		*latestDesired != baseDesired || *latestMin != baseMin || *latestMax != baseMax
}

func ShowVersion() {
	r, _ := regexp.Compile(`v[0-9]\.[0-9]+\.[0-9]+`)
	versionInfo := r.FindString(constants.GitVersion)
	fmt.Println(constants.AppName, versionInfo)
	fmt.Println(" Git Commit:", constants.GitVersion)
	fmt.Println(" Build with:", constants.GoVersion)
	fmt.Println(" Build time:", constants.BuildTime)
}

func printRequestDetails(clusterName, nodegroupName string, desiredSize, minSize, maxSize int32, sentAt time.Time) {
	fmt.Println("Request details:")
	fmt.Printf("  Cluster:      %s\n", clusterName)
	fmt.Printf("  Node Group:   %s\n", nodegroupName)
	fmt.Printf("  Desired Size: %d\n", desiredSize)
	fmt.Printf("  Min Size:     %d\n", minSize)
	fmt.Printf("  Max Size:     %d\n", maxSize)
	fmt.Printf("  Sent At:      %s\n", sentAt.Format(time.RFC3339))
}

func printSizeChange(clusterName, nodegroupName string, currentDesired, currentMin, currentMax, desiredSize, minSize, maxSize int32) {
	fmt.Printf("  Cluster:      %s\n", clusterName)
	fmt.Printf("  Node Group:   %s\n", nodegroupName)
	fmt.Printf("  Desired Size: %d -> %d\n", currentDesired, desiredSize)
	fmt.Printf("  Min Size:     %d -> %d\n", currentMin, minSize)
	fmt.Printf("  Max Size:     %d -> %d\n", currentMax, maxSize)
}

func confirmSizeChange() (bool, error) {
	var confirmed bool
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Apply this change?").
				Value(&confirmed),
		),
	)
	if err := form.Run(); err != nil {
		return false, fmt.Errorf("confirmation form error: %w", err)
	}
	return confirmed, nil
}

func reportChange(opts Options, nodegroupName string, currentDesired, currentMin, currentMax, desiredSize, minSize, maxSize int32, applied bool, sentAt time.Time) error {
	sentAt = sentAt.UTC()
	if opts.Output == "json" {
		report := changeReport{
			Cluster:        opts.ClusterName,
			Nodegroup:      nodegroupName,
			CurrentDesired: currentDesired,
			CurrentMin:     currentMin,
			CurrentMax:     currentMax,
			DesiredSize:    desiredSize,
			MinSize:        minSize,
			MaxSize:        maxSize,
			Applied:        applied,
		}
		if applied {
			report.SentAt = sentAt.Format(time.RFC3339)
		}
		return printJSON(report)
	}

	if applied {
		printRequestDetails(opts.ClusterName, nodegroupName, desiredSize, minSize, maxSize, sentAt)
	} else {
		fmt.Println("Dry run, no change applied:")
		printSizeChange(opts.ClusterName, nodegroupName, currentDesired, currentMin, currentMax, desiredSize, minSize, maxSize)
	}
	return nil
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

	desired, err := utils.ParseInt32(desiredSize)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid desired size: %w", err)
	}

	min, err := utils.ParseInt32(minSize)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid min size: %w", err)
	}

	max, err := utils.ParseInt32(maxSize)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid max size: %w", err)
	}

	return desired, min, max, nil
}

func resolveNodegroupName(opts Options, nodegroups []string) (string, error) {
	if opts.NodegroupName != "" {
		if !slices.Contains(nodegroups, opts.NodegroupName) {
			return "", fmt.Errorf("nodegroup %q not found in cluster %q", opts.NodegroupName, opts.ClusterName)
		}
		return opts.NodegroupName, nil
	}
	return nodegroupsForm(nodegroups)
}

func resolveNodegroupSize(opts Options) (int32, int32, int32, error) {
	var desiredSize, minSize, maxSize int32
	var err error

	if opts.DesiredSize != nil && opts.MinSize != nil && opts.MaxSize != nil {
		desiredSize, minSize, maxSize = *opts.DesiredSize, *opts.MinSize, *opts.MaxSize
	} else {
		desiredSize, minSize, maxSize, err = nodegroupSizeForm()
		if err != nil {
			return 0, 0, 0, err
		}
	}

	if err := utils.ValidateNodegroupSize(desiredSize, minSize, maxSize); err != nil {
		return 0, 0, 0, err
	}
	return desiredSize, minSize, maxSize, nil
}

func selfManagedNodeGroupWorkflow(asgClient *autoscaling.Client, opts Options) error {
	clusterName := opts.ClusterName
	nodegroups, err := asgwrapper.GetAutoScalingGroupsByClusterName(asgClient, clusterName)
	if err != nil {
		return err
	}
	if len(nodegroups) == 0 {
		return printStatus(opts, "no nodegroup found")
	}
	nodegroupName, err := resolveNodegroupName(opts, nodegroups)
	if err != nil {
		return err
	}

	desiredSize, minSize, maxSize, err := resolveNodegroupSize(opts)
	if err != nil {
		return err
	}

	describeAutoScalingGroupsOutput, err := asgwrapper.DescribeAutoScalingGroupsByNodegroupName(asgClient, nodegroupName)
	if err != nil {
		return err
	}

	if len(describeAutoScalingGroupsOutput.AutoScalingGroups) == 0 {
		return fmt.Errorf("auto scaling group %s not found", nodegroupName)
	}

	group := describeAutoScalingGroupsOutput.AutoScalingGroups[0]
	if group.DesiredCapacity == nil || group.MinSize == nil || group.MaxSize == nil {
		return fmt.Errorf("auto scaling group %s has nil capacity values", nodegroupName)
	}
	currentDesiredCapacity := *group.DesiredCapacity
	currentMinSize := *group.MinSize
	currentMaxSize := *group.MaxSize

	if currentDesiredCapacity == desiredSize && currentMinSize == minSize && currentMaxSize == maxSize {
		return printStatus(opts, "no change required, target node group size have no difference")
	}

	if opts.DryRun {
		return reportChange(opts, nodegroupName, currentDesiredCapacity, currentMinSize, currentMaxSize, desiredSize, minSize, maxSize, false, time.Time{})
	}

	if !opts.Yes {
		fmt.Println("About to apply:")
		printSizeChange(clusterName, nodegroupName, currentDesiredCapacity, currentMinSize, currentMaxSize, desiredSize, minSize, maxSize)
		confirmed, err := confirmSizeChange()
		if err != nil {
			return err
		}
		if !confirmed {
			return printStatus(opts, "aborted, no change applied")
		}

		// the auto scaling group may have been modified while waiting for
		// confirmation, re-read and make sure it still matches what was shown
		latestOutput, err := asgwrapper.DescribeAutoScalingGroupsByNodegroupName(asgClient, nodegroupName)
		if err != nil {
			return err
		}
		if len(latestOutput.AutoScalingGroups) == 0 {
			return fmt.Errorf("auto scaling group %s not found", nodegroupName)
		}
		latest := latestOutput.AutoScalingGroups[0]
		if scalingConfigDrifted(currentDesiredCapacity, currentMinSize, currentMaxSize, latest.DesiredCapacity, latest.MinSize, latest.MaxSize) {
			return fmt.Errorf("auto scaling group %s was modified while waiting for confirmation, please re-run", nodegroupName)
		}
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

	return reportChange(opts, nodegroupName, currentDesiredCapacity, currentMinSize, currentMaxSize, desiredSize, minSize, maxSize, true, time.Now())
}

func managedNodeGroupWorkflow(eksClient *eks.Client, opts Options) error {
	clusterName := opts.ClusterName
	nodegroups, err := ekswrapper.ListNodegroups(eksClient, clusterName)
	if err != nil {
		return err
	}
	if len(nodegroups) == 0 {
		return printStatus(opts, "no nodegroup found")
	}
	nodegroupName, err := resolveNodegroupName(opts, nodegroups)
	if err != nil {
		return err
	}

	desiredSize, minSize, maxSize, err := resolveNodegroupSize(opts)
	if err != nil {
		return err
	}

	scalingConfig, err := ekswrapper.GetNodegroupScalingConfig(eksClient, clusterName, nodegroupName)
	if err != nil {
		return err
	}
	if scalingConfig.DesiredSize == nil || scalingConfig.MinSize == nil || scalingConfig.MaxSize == nil {
		return fmt.Errorf("nodegroup %s has nil scaling config values", nodegroupName)
	}
	if *scalingConfig.DesiredSize == desiredSize && *scalingConfig.MinSize == minSize && *scalingConfig.MaxSize == maxSize {
		return printStatus(opts, "no change required, target node group size have no difference")
	}

	if opts.DryRun {
		return reportChange(opts, nodegroupName, *scalingConfig.DesiredSize, *scalingConfig.MinSize, *scalingConfig.MaxSize, desiredSize, minSize, maxSize, false, time.Time{})
	}

	if !opts.Yes {
		fmt.Println("About to apply:")
		printSizeChange(clusterName, nodegroupName, *scalingConfig.DesiredSize, *scalingConfig.MinSize, *scalingConfig.MaxSize, desiredSize, minSize, maxSize)
		confirmed, err := confirmSizeChange()
		if err != nil {
			return err
		}
		if !confirmed {
			return printStatus(opts, "aborted, no change applied")
		}

		// the nodegroup may have been modified while waiting for confirmation,
		// re-read and make sure it still matches what was shown
		latest, err := ekswrapper.GetNodegroupScalingConfig(eksClient, clusterName, nodegroupName)
		if err != nil {
			return err
		}
		if scalingConfigDrifted(*scalingConfig.DesiredSize, *scalingConfig.MinSize, *scalingConfig.MaxSize, latest.DesiredSize, latest.MinSize, latest.MaxSize) {
			return fmt.Errorf("nodegroup %s was modified while waiting for confirmation, please re-run", nodegroupName)
		}
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
	if result.Update == nil || result.Update.CreatedAt == nil {
		return fmt.Errorf("invalid update response for nodegroup %s", nodegroupName)
	}

	return reportChange(opts, nodegroupName, *scalingConfig.DesiredSize, *scalingConfig.MinSize, *scalingConfig.MaxSize, desiredSize, minSize, maxSize, true, *result.Update.CreatedAt)
}

func Entry(opts Options) error {
	if err := opts.Validate(); err != nil {
		return err
	}

	if err := ekswrapper.ValidateCredentials(opts.Region, opts.Profile); err != nil {
		return fmt.Errorf("credential validation failed: %w", err)
	}

	var eksClient *eks.Client
	var err error

	// self-managed node groups with an explicit cluster name need no EKS API
	// access, the cluster tag filter on auto scaling groups does the scoping
	if opts.NodeGroupType != "self-managed" || opts.ClusterName == "" {
		eksClient, err = ekswrapper.GetEksClient(opts.Region, opts.Profile)
		if err != nil {
			return err
		}

		clusters, err := ekswrapper.ListClusters(eksClient)
		if err != nil {
			return err
		}
		if len(clusters) == 0 {
			return printStatus(opts, "no cluster found")
		}

		if opts.ClusterName != "" {
			if !slices.Contains(clusters, opts.ClusterName) {
				return fmt.Errorf("cluster %q not found in region %q", opts.ClusterName, opts.Region)
			}
		} else {
			opts.ClusterName, err = clustersForm(clusters)
			if err != nil {
				return err
			}
		}
	}

	var nodeGroupType string
	switch opts.NodeGroupType {
	case "managed":
		nodeGroupType = constants.NodeGroupTypes[constants.Managed]
	case "self-managed":
		nodeGroupType = constants.NodeGroupTypes[constants.SelfManaged]
	default:
		nodeGroupType, err = nodeGroupTypeForm()
		if err != nil {
			return err
		}
	}

	if nodeGroupType == constants.NodeGroupTypes[constants.SelfManaged] {
		asgClient, err := asgwrapper.GetAsgClient(opts.Region, opts.Profile)
		if err != nil {
			return err
		}

		if err := selfManagedNodeGroupWorkflow(asgClient, opts); err != nil {
			return err
		}
	} else {
		if err := managedNodeGroupWorkflow(eksClient, opts); err != nil {
			return err
		}
	}
	return nil
}
