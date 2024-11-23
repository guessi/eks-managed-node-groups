package ui

import (
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/guessi/eks-managed-node-groups/pkg/constants"
	eksclient "github.com/guessi/eks-managed-node-groups/pkg/eks"
	"github.com/guessi/eks-managed-node-groups/pkg/utils"
)

func ShowVersion() {
	r, _ := regexp.Compile(`v[0-9]\.[0-9]+\.[0-9]+`)
	versionInfo := r.FindString(constants.GitVersion)
	fmt.Println("eks-managed-node-groups", versionInfo)
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

func Entry() error {
	client := eksclient.GetEksClient()

	clusters := eksclient.ListClusters(client)
	if len(clusters) == 0 {
		fmt.Println("No cluster found.")
		return nil
	}
	clusterName := clustersForm(clusters)

	nodegroups := eksclient.ListNodegroups(client, clusterName)
	if len(nodegroups) == 0 {
		fmt.Println("No nodegroups found.")
		return nil
	}
	nodegroupName := nodegroupsForm(nodegroups)

	desiredSize, minSize, maxSize := nodegroupSizeForm()

	currentDesireSize, currentMinSize, currentMaxSize := eksclient.GetNodegroupScalingConfig(client, clusterName, nodegroupName)

	if isValid := utils.ValidateNodegroupSize(desiredSize, minSize, maxSize); isValid {
		if currentDesireSize == desiredSize && currentMinSize == minSize && currentMaxSize == maxSize {
			fmt.Println("No change required, target node group size have no difference.")
			return nil
		}
		result, err := eksclient.UpdateNodegroupConfig(client, clusterName, nodegroupName, desiredSize, minSize, maxSize)
		if err != nil {
			return err
		}
		fmt.Printf("Request details: {desiredSize: %d, minSize: %d, maxSize: %d}\n", desiredSize, minSize, maxSize)
		fmt.Printf("Request sent at: %s\n", result.Update.CreatedAt.Format(time.RFC3339))
	}
	return nil
}
