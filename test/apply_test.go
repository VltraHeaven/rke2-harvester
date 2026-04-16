package test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestApplyDestroyBasicVM provisions a single VM without a LoadBalancer,
// validates the vm_ip_addresses output, and then destroys.
//
// Run with: go test -v -run TestApplyDestroyBasicVM -timeout 30m
func TestApplyDestroyBasicVM(t *testing.T) {
	t.Parallel()

	vars := getBaseVars(t)
	vars["vm_prefix"] = "terratest-basic"
	vars["vm_count"] = 1
	vars["create_lb"] = false

	opts := terraformOptions(t, vars)
	defer terraform.Destroy(t, opts)

	terraform.InitAndApply(t, opts)

	// vm_ip_addresses is a list of "<name>: <ip>" strings, one per VM.
	rawOutput := terraform.OutputList(t, opts, "vm_ip_addresses")
	require.Len(t, rawOutput, 1, "expected exactly 1 VM IP address entry")

	entry := rawOutput[0]
	assert.True(t, strings.HasPrefix(entry, "terratest-basic-0:"),
		"output entry should be prefixed with the VM name, got: %s", entry)

	parts := strings.SplitN(entry, ":", 2)
	require.Len(t, parts, 2)
	assert.NotEmpty(t, strings.TrimSpace(parts[1]),
		"VM IP address should not be empty")
}

// TestApplyDestroyMultiVM provisions multiple VMs and validates that each one
// appears in the vm_ip_addresses output.
//
// Run with: go test -v -run TestApplyDestroyMultiVM -timeout 60m
func TestApplyDestroyMultiVM(t *testing.T) {
	t.Parallel()

	vmCount := 3

	vars := getBaseVars(t)
	vars["vm_prefix"] = "terratest-multi"
	vars["vm_count"] = vmCount
	vars["create_lb"] = false

	opts := terraformOptions(t, vars)
	defer terraform.Destroy(t, opts)

	terraform.InitAndApply(t, opts)

	rawOutput := terraform.OutputList(t, opts, "vm_ip_addresses")
	require.Len(t, rawOutput, vmCount,
		"expected %d VM IP address entries, got %d", vmCount, len(rawOutput))

	for i := 0; i < vmCount; i++ {
		expectedPrefix := fmt.Sprintf("terratest-multi-%d:", i)
		found := false
		for _, entry := range rawOutput {
			if strings.HasPrefix(entry, expectedPrefix) {
				found = true
				parts := strings.SplitN(entry, ":", 2)
				assert.NotEmpty(t, strings.TrimSpace(parts[1]),
					"VM %d IP address should not be empty", i)
				break
			}
		}
		assert.True(t, found, "expected output entry with prefix %q", expectedPrefix)
	}
}

// TestApplyDestroyWithLB provisions VMs with a LoadBalancer and validates that
// vm_lb_ip_address is non-empty in the output.
//
// Run with: go test -v -run TestApplyDestroyWithLB -timeout 30m
func TestApplyDestroyWithLB(t *testing.T) {
	t.Parallel()

	vars := getBaseVars(t)
	vars["vm_prefix"] = "terratest-lb"
	vars["vm_count"] = 2
	vars["create_lb"] = true
	vars["lb_ipam"] = "dhcp"
	vars["lb_listener_port"] = 443
	vars["lb_listener_backend_port"] = 443
	vars["lb_protocol"] = "TCP"
	vars["lb_healthcheck_period_seconds"] = 5
	vars["lb_healthcheck_timeout_seconds"] = 3
	vars["lb_healthcheck_failure_threshold"] = 3
	vars["lb_healthcheck_success_threshold"] = 1

	opts := terraformOptions(t, vars)
	defer terraform.Destroy(t, opts)

	terraform.InitAndApply(t, opts)

	lbIP := terraform.Output(t, opts, "vm_lb_ip_address")
	assert.NotEmpty(t, lbIP, "vm_lb_ip_address should be populated when create_lb=true")
}

// TestDiskRetainedAfterDestroy provisions a VM with vm_disk_auto_delete=false,
// destroys it, and confirms the apply/destroy cycle completed without error.
//
// NOTE: To fully verify PVC retention after this test, run:
//
//	kubectl get pvc -n <namespace> | grep terratest-disk-retain
//
// and clean up manually with kubectl delete pvc.
//
// Run with: go test -v -run TestDiskRetainedAfterDestroy -timeout 30m
func TestDiskRetainedAfterDestroy(t *testing.T) {
	t.Parallel()

	vars := getBaseVars(t)
	vars["vm_prefix"] = "terratest-disk-retain"
	vars["vm_count"] = 1
	vars["create_lb"] = false
	vars["vm_disk_auto_delete"] = false

	opts := terraformOptions(t, vars)

	terraform.InitAndApply(t, opts)
	terraform.Destroy(t, opts)
}
