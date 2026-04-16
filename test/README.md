# Terratest Suite — terraform-harvester

Tests live in the `test/` directory and are written using [Terratest](https://terratest.gruntwork.io/).

## Prerequisites

- Go >= 1.21
- `terraform` >= 0.13 in `$PATH`
- `kubectl` in `$PATH` (required by the module's `local-exec` IP-wait provisioner)
- Access to a live Harvester cluster for apply tests

## Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `HARVESTER_KUBECONFIG` | Yes | Path to the Harvester kubeconfig file |
| `HARVESTER_NAMESPACE` | No | Namespace to deploy into (default: `default`) |
| `HARVESTER_NETWORK` | Yes | Name of the Harvester VM network to attach VMs to |
| `HARVESTER_NETWORK_NAMESPACE` | No | Namespace of the network (defaults to `HARVESTER_NAMESPACE`) |
| `HARVESTER_IMAGE_DISPLAY_NAME` | Yes* | Display name of an existing VM image (required when `download_image=false`) |

## Running the Tests

### Install dependencies

```bash
cd test/
go mod tidy
```

### Plan-only tests (fast, no infrastructure)

These validate HCL correctness. They require provider access to resolve data sources but do not create VMs.

```bash
export HARVESTER_KUBECONFIG=/path/to/harvester.yaml
export HARVESTER_NETWORK=vlan100
export HARVESTER_IMAGE_DISPLAY_NAME="noble-server-cloudimg-amd64.img"

go test -v -run "TestPlan" -timeout 10m ./...
```

### Full apply/destroy tests (slow, creates real infrastructure)

```bash
go test -v -run "TestApply" -timeout 60m ./...
```

### Run a single test

```bash
go test -v -run TestPlanNoLB_OutputSafe -timeout 10m ./...
```

### Run everything

```bash
go test -v -timeout 90m ./...
```

## Test Inventory

### `plan_test.go` — Plan-only tests (no apply)

| Test | Description |
|------|-------------|
| `TestPlanNoLB` | Plan must succeed with `create_lb=false` |
| `TestPlanWithLB` | Plan must succeed with `create_lb=true` |


### `apply_test.go` — Full apply/destroy tests

| Test | Description |
|------|-------------|
| `TestApplyDestroyBasicVM` | Single VM apply; validates `vm_ip_addresses` output format |
| `TestApplyDestroyMultiVM` | Multi-VM apply; validates one output entry per VM |
| `TestApplyDestroyWithLB` | Apply with LB; validates `vm_lb_ip_address` is non-empty |
| `TestDiskRetainedAfterDestroy` | Destroy with `vm_disk_auto_delete=false` completes without error |

### `image_test.go` — Image download tests

| Test | Description |
|------|-------------|
| `TestApplyDownloadImage` | Full apply downloading a fresh image and booting a VM from it |
| `TestPlanDownloadImageEnabled` | Plan must succeed with `download_image=true` and a valid `new_image` object |
| `TestPlanDownloadImageDisabled` | `harvester_image` resource must be absent from plan when `download_image=false` |

## Notes

- All apply tests use `defer terraform.Destroy()` to ensure cleanup even on failure.
- VM name prefixes are unique per test to avoid resource conflicts when tests run in parallel.
- `TestDiskRetainedAfterDestroy` intentionally leaves a PVC behind. After the test, manually verify with:
  ```bash
  kubectl get pvc -n <namespace> | grep terratest-disk-retain
  ```
  and clean up with:
  ```bash
  kubectl delete pvc -n <namespace> <pvc-name>
  ```
