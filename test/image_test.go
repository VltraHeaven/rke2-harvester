package test

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
)

// TestPlanDownloadImageEnabled verifies that setting download_image=true with a
// valid new_image object produces a successful plan. No apply is performed.
func TestPlanDownloadImageEnabled(t *testing.T) {
	t.Parallel()

	vars := getBaseVars(t)
	vars["vm_prefix"] = "test-dl-image"
	vars["download_image"] = true
	vars["image_namespace"] = vars["namespace"]
	vars["new_image"] = map[string]interface{}{
		"name":         "ubuntu24-test",
		"display_name": "noble-server-cloudimg-amd64.img",
		"source_type":  "download",
		"url":          "https://cloud-images.ubuntu.com/noble/current/noble-server-cloudimg-amd64.img",
	}

	opts := terraformOptions(t, vars)

	exitCode := terraform.InitAndPlanWithExitCode(t, opts)
	assert.Contains(t, []int{0, 2}, exitCode,
		"plan should succeed with download_image=true and a valid new_image object")
}

// TestPlanDownloadImageDisabled verifies that setting download_image=false and
// providing image_display_name produces a valid plan, and that no
// harvester_image resource is included in the planned changes.
func TestPlanDownloadImageDisabled(t *testing.T) {
	t.Parallel()

	vars := getBaseVars(t)
	vars["vm_prefix"] = "test-nodl-image"
	vars["download_image"] = false

	opts := terraformOptions(t, vars)

	planStruct := terraform.InitAndPlanAndShowWithStructNoLogTempPlanFile(t, opts)

	for addr := range planStruct.ResourcePlannedValuesMap {
		assert.NotContains(t, addr, "harvester_image.new_image",
			"harvester_image.new_image should not be planned when download_image=false")
	}
}

// TestApplyDownloadImage performs a full apply with download_image=true,
// provisions a VM from the freshly downloaded image, then destroys everything.
//
// Run with: go test -v -run TestApplyDownloadImage -timeout 60m
func TestApplyDownloadImage(t *testing.T) {
	t.Parallel()

	vars := getBaseVars(t)
	vars["vm_prefix"] = "terratest-dl-image"
	vars["vm_count"] = 1
	vars["download_image"] = true
	vars["image_namespace"] = vars["namespace"]
	vars["new_image"] = map[string]interface{}{
		"name":         "ubuntu24-terratest",
		"display_name": "noble-server-cloudimg-amd64.img",
		"source_type":  "download",
		"url":          "https://cloud-images.ubuntu.com/noble/current/noble-server-cloudimg-amd64.img",
	}

	opts := terraformOptions(t, vars)
	defer terraform.Destroy(t, opts)

	terraform.InitAndApply(t, opts)

	// If the VM came up with an IP, the image downloaded and booted correctly.
	rawOutput := terraform.OutputList(t, opts, "vm_ip_addresses")
	assert.Len(t, rawOutput, 1, "expected 1 VM IP address after apply with downloaded image")
}
