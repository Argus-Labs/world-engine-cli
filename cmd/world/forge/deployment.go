package forge

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/rotisserie/eris"

	"pkg.world.dev/world-cli/common/globalconfig"
)

var statusFailRegEx = regexp.MustCompile(`[^a-zA-Z0-9\. ]+`)

// Deploy a project
func deploy(ctx context.Context) error {
	globalConfig, err := globalconfig.GetGlobalConfig()
	if err != nil {
		return eris.Wrap(err, "Failed to get global config")
	}

	projectID := globalConfig.ProjectID
	organizationID := globalConfig.OrganizationID

	if organizationID == "" {
		printNoSelectedOrganization()
		return nil
	}

	if projectID == "" {
		printNoSelectedProject()
		return nil
	}

	// Get organization details
	org, err := getSelectedOrganization(ctx)
	if err != nil {
		return eris.Wrap(err, "Failed to get organization details")
	}

	// Get project details
	prj, err := getSelectedProject(ctx)
	if err != nil {
		return eris.Wrap(err, "Failed to get project details")
	}

	fmt.Println("Deployment Details")
	fmt.Println("-----------------")
	fmt.Printf("Organization: %s\n", org.Name)
	fmt.Printf("Org Slug:     %s\n", org.Slug)
	fmt.Printf("Project:      %s\n", prj.Name)
	fmt.Printf("Project Slug: %s\n", prj.Slug)
	fmt.Printf("Repository:   %s\n\n", prj.RepoURL)

	deployURL := fmt.Sprintf("%s/api/organization/%s/project/%s/deploy", baseURL, organizationID, projectID)
	_, err = sendRequest(ctx, http.MethodPost, deployURL, nil)
	if err != nil {
		return eris.Wrap(err, "Failed to deploy project")
	}

	fmt.Println("\n✨ Your deployment is being processed! ✨")
	fmt.Println("\nTo check the status of your deployment, run:")
	fmt.Println("  $ 'world forge deployment status'")

	return nil
}

// Destroy a project
func destroy(ctx context.Context) error {
	globalConfig, err := globalconfig.GetGlobalConfig()
	if err != nil {
		return eris.Wrap(err, "Failed to get global config")
	}

	projectID := globalConfig.ProjectID
	organizationID := globalConfig.OrganizationID

	if organizationID == "" {
		printNoSelectedOrganization()
		return nil
	}

	if projectID == "" {
		printNoSelectedProject()
		return nil
	}

	// Get organization details
	org, err := getSelectedOrganization(ctx)
	if err != nil {
		return eris.Wrap(err, "Failed to get organization details")
	}

	// Get project details
	prj, err := getSelectedProject(ctx)
	if err != nil {
		return eris.Wrap(err, "Failed to get project details")
	}

	fmt.Println("Project Details")
	fmt.Println("-----------------")
	fmt.Printf("Organization: %s\n", org.Name)
	fmt.Printf("Org Slug:     %s\n", org.Slug)
	fmt.Printf("Project:      %s\n", prj.Name)
	fmt.Printf("Project Slug: %s\n", prj.Slug)
	fmt.Printf("Repository:   %s\n\n", prj.RepoURL)

	fmt.Print("Are you sure you want to destroy this project? (y/N): ")
	response, err := getInput()
	if err != nil {
		return eris.Wrap(err, "Failed to read response")
	}

	response = strings.ToLower(strings.TrimSpace(response))
	if response != "y" {
		fmt.Println("Destroy cancelled")
		return nil
	}

	destroyURL := fmt.Sprintf("%s/api/organization/%s/project/%s/destroy", baseURL, organizationID, projectID)
	_, err = sendRequest(ctx, http.MethodPost, destroyURL, nil)
	if err != nil {
		return eris.Wrap(err, "Failed to destroy project")
	}

	fmt.Println("\n🗑️  Your destroy request is being processed!")
	fmt.Println("\nTo check the status of your destroy request, run:")
	fmt.Println("  $ 'world forge deployment status'")

	return nil
}

//nolint:funlen, gocognit, gocyclo, cyclop // this is actually a straightforward function with a lot of error handling
func status(ctx context.Context) error {
	globalConfig, err := globalconfig.GetGlobalConfig()
	if err != nil {
		return eris.Wrap(err, "Failed to get global config")
	}
	projectID := globalConfig.ProjectID
	if projectID == "" {
		printNoSelectedProject()
		return nil
	}
	// Get project details
	prj, err := getSelectedProject(ctx)
	if err != nil {
		return eris.Wrap(err, "Failed to get project details")
	}

	statusURL := fmt.Sprintf("%s/api/deployment/%s", baseURL, projectID)
	result, err := sendRequest(ctx, http.MethodGet, statusURL, nil)
	if err != nil {
		return eris.Wrap(err, "Failed to get deployment status")
	}
	var response map[string]any
	err = json.Unmarshal(result, &response)
	if err != nil {
		return eris.Wrap(err, "Failed to unmarshal deployment response")
	}
	var data map[string]any
	if response["data"] != nil {
		// data = null is returned when there are no deployments, so we have to check for that before we
		// try to cast the response into a json object map, since this is not an error but the cast would
		// fail
		var ok bool
		data, ok = response["data"].(map[string]any)
		if !ok {
			return eris.New("Failed to unmarshal deployment data")
		}
	}
	fmt.Println("Deployment Status")
	fmt.Println("-----------------")
	fmt.Printf("Project:      %s\n", prj.Name)
	fmt.Printf("Project Slug: %s\n", prj.Slug)
	fmt.Printf("Repository:   %s\n", prj.RepoURL)
	if data == nil {
		fmt.Printf("\n** Project has not been deployed **\n")
		return nil
	}
	if data["project_id"] != projectID {
		return eris.Errorf("Deployment status does not match project id %s", projectID)
	}
	if data["type"] != "deploy" {
		return eris.Errorf("Deployment status does not match type %s", data["type"])
	}
	executorID, ok := data["executor_id"].(string)
	if !ok {
		return eris.New("Failed to unmarshal deployment executor_id")
	}
	executionTimeStr, ok := data["execution_time"].(string)
	if !ok {
		return eris.New("Failed to unmarshal deployment execution_time")
	}
	dt, dte := time.Parse(time.RFC3339, executionTimeStr)
	if dte != nil {
		return eris.Wrapf(dte, "Failed to parse deployment execution_time %s", dt)
	}
	bnf, ok := data["build_number"].(float64)
	if !ok {
		return eris.New("Failed to unmarshal deployment build_number")
	}
	buildNumber := int(bnf)
	buildTimeStr, ok := data["build_time"].(string)
	if !ok {
		return eris.New("Failed to unmarshal deployment build_time")
	}
	bt, bte := time.Parse(time.RFC3339, buildTimeStr)
	if bte != nil {
		return eris.Wrapf(bte, "Failed to parse deployment build_time %s", bt)
	}
	buildState, ok := data["build_state"].(string)
	if !ok {
		return eris.New("Failed to unmarshal deployment build_state")
	}
	if buildState != "finished" {
		fmt.Printf("Build:        #%d started %s by %s - %s\n", buildNumber, dt.Format(time.RFC822), executorID, buildState)
		return nil
	}
	fmt.Printf("Build:        #%d on %s by %s\n", buildNumber, dt.Format(time.RFC822), executorID)
	fmt.Print("Health:       ")

	// fmt.Println()
	//	fmt.Println(string(result))

	healthURL := fmt.Sprintf("%s/api/health/%s", baseURL, projectID)
	result, err = sendRequest(ctx, http.MethodGet, healthURL, nil)
	if err != nil {
		return eris.Wrap(err, "Failed to get health")
	}
	err = json.Unmarshal(result, &response)
	if err != nil {
		return eris.Wrap(err, "Failed to unmarshal health response")
	}
	if response["data"] == nil {
		return eris.New("Failed to unmarshal health data")
	}
	instances, ok := response["data"].([]any)
	if !ok {
		return eris.New("Failed to unmarshal deployment status")
	}
	if len(instances) == 0 {
		fmt.Println("** No deployed instances found **")
		return nil
	}
	fmt.Printf("(%d deployed instances)\n", len(instances))
	currRegion := ""
	for _, instance := range instances {
		info, ok := instance.(map[string]any)
		if !ok {
			return eris.Errorf("Failed to unmarshal deployment instance %d info", instance)
		}
		region, ok := info["region"].(string)
		if !ok {
			return eris.Errorf("Failed to unmarshal deployment instance %d region", instance)
		}
		instancef, ok := info["instance"].(float64)
		if !ok {
			return eris.Errorf("Failed to unmarshal deployment instance %d instance number", instance)
		}
		instanceNum := int(instancef)
		cardinalInfo, ok := info["cardinal"].(map[string]any)
		if !ok {
			return eris.Errorf("Failed to unmarshal deployment instance %d cardinal data", instance)
		}
		nakamaInfo, ok := info["nakama"].(map[string]any)
		if !ok {
			return eris.Errorf("Failed to unmarshal deployment instance %d nakama data", instance)
		}
		cardinalURL, ok := cardinalInfo["url"].(string)
		if !ok {
			return eris.Errorf("Failed to unmarshal deployment instance %d cardinal url", instance)
		}
		cardinalHost := strings.Split(cardinalURL, "/")[2]
		cardinalOK, ok := cardinalInfo["ok"].(bool)
		if !ok {
			return eris.Errorf("Failed to unmarshal deployment instance %d cardinal ok flag", instance)
		}
		cardinalResultCodef, ok := cardinalInfo["result_code"].(float64)
		if !ok {
			return eris.Errorf("Failed to unmarshal deployment instance %d cardinal result_code", instance)
		}
		cardinalResultCode := int(cardinalResultCodef)
		cardinalResultStr, ok := cardinalInfo["result_str"].(string)
		if !ok {
			return eris.Errorf("Failed to unmarshal deployment instance %d cardinal result_str", instance)
		}
		nakamaURL, ok := nakamaInfo["url"].(string)
		if !ok {
			return eris.Errorf("Failed to unmarshal deployment instance %d nakama url", instance)
		}
		nakamaHost := strings.Split(nakamaURL, "/")[2]
		nakamaOK, ok := nakamaInfo["ok"].(bool)
		if !ok {
			return eris.Errorf("Failed to unmarshal deployment instance %d nakama ok", instance)
		}
		nakamaResultCodef, ok := nakamaInfo["result_code"].(float64)
		if !ok {
			return eris.Errorf("Failed to unmarshal deployment instance %d result_code", instance)
		}
		nakamaResultCode := int(nakamaResultCodef)
		nakamaResultStr, ok := nakamaInfo["result_str"].(string)
		if !ok {
			return eris.Errorf("Failed to unmarshal deployment instance %d result_str", instance)
		}
		if region != currRegion {
			currRegion = region
			fmt.Printf("• %s\n", currRegion)
		}
		fmt.Printf("  %d)", instanceNum)
		fmt.Printf("\tCardinal: %s - ", cardinalHost)
		switch {
		case cardinalOK:
			fmt.Print("OK\n")
		case cardinalResultCode == 0:
			fmt.Printf("FAIL %s\n", statusFailRegEx.ReplaceAllString(cardinalResultStr, ""))
		default:
			fmt.Printf("FAIL %d %s\n", cardinalResultCode, statusFailRegEx.ReplaceAllString(cardinalResultStr, ""))
		}
		fmt.Printf("\tNakama:   %s - ", nakamaHost)
		switch {
		case nakamaOK:
			fmt.Print("OK\n")
		case nakamaResultCode == 0:
			fmt.Printf("FAIL %s\n", statusFailRegEx.ReplaceAllString(nakamaResultStr, ""))
		default:
			fmt.Printf("FAIL %d %s\n", nakamaResultCode, statusFailRegEx.ReplaceAllString(nakamaResultStr, ""))
		}
	}
	// fmt.Println()
	// fmt.Println(string(result))

	return nil
}
