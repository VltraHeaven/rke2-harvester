# terraform-harvester

A Terraform module for provisioning virtual machines on [Harvester HCI](https://harvesterhci.io/), SUSE's open-source hyperconverged infrastructure platform built on Kubernetes and KubeVirt. The module handles VM creation with cloud-init support, optional image downloading, and optional LoadBalancer provisioning.

## Features

- Deploy one or more identical VMs from an existing or newly downloaded VM image
- Full cloud-init `user-data` and `network-data` support via Kubernetes secrets
- Optional Harvester LoadBalancer resource with configurable health checks
- EFI boot enabled by default; `virtio` disk bus for optimal performance
- Post-provisioning IP readiness probe via `local-exec`

---

## Requirements

| Name | Version |
|------|---------|
| Terraform | >= 1.3 |
| [harvester/harvester](https://registry.terraform.io/providers/harvester/harvester/latest) | >= 0.6 |
| `kubectl` | Available in `$PATH` (used by the IP-wait `local-exec` provisioner) |

## Provider Configuration

The module uses the `harvester` provider, configured with a kubeconfig file and optional context name:

```hcl
provider "harvester" {
  kubeconfig = var.kconfig
  context    = var.kcontext
}
```

---

## Usage

### Minimal Example — Use an existing image

```hcl
module "vms" {
  source = "github.com/VltraHeaven/terraform-harvester"

  kconfig                 = "/home/user/.kube/harvester.yaml"
  namespace               = "default"
  vm_prefix               = "webserver"
  vm_count                = 3
  harvester_net           = "vlan100"
  harvester_net_namespace = "default"
  image_name              = "ubuntu24"
  image_namespace         = "default"
  image_storageclass      = "harvester-longhorn"
}
```

### Full Example — Download a new image and provision a LoadBalancer

```hcl
module "vms" {
  source = "github.com/VltraHeaven/terraform-harvester"

  kconfig                 = "/home/user/.kube/harvester.yaml"
  kcontext                = "local"
  namespace               = "default"

  download_image = true
  new_image = {
    name         = "ubuntu24"
    display_name = "noble-server-cloudimg-amd64.img"
    source_type  = "download"
    url          = "https://cloud-images.ubuntu.com/noble/current/noble-server-cloudimg-amd64.img"
  }

  vm_prefix      = "k8s-node"
  vm_count       = 3
  vm_cpu         = 4
  vm_memory      = "8Gi"
  vm_disksize    = "50Gi"
  vm_description = "Kubernetes worker nodes"
  vm_labels      = { env = "prod", role = "worker" }
  ssh_user       = "ubuntu"

  harvester_net           = "vlan100"
  harvester_net_namespace = "default"

  cloud_config_user_data = <<-EOF
    #cloud-config
    package_update: true
    packages:
      - qemu-guest-agent
    runcmd:
      - [systemctl, enable, qemu-guest-agent.service]
      - [systemctl, start, qemu-guest-agent.service]
    users:
      - name: ubuntu
        sudo: ALL=(ALL) NOPASSWD:ALL
        ssh_authorized_keys:
          - ssh-ed25519 AAAA... your-key-here
  EOF

  create_lb                        = true
  lb_ipam                          = "pool"
  lb_listener_port                 = 6443
  lb_listener_backend_port         = 6443
  lb_protocol                      = "TCP"
  lb_healthcheck_period_seconds    = 5
  lb_healthcheck_timeout_seconds   = 3
  lb_healthcheck_failure_threshold = 3
  lb_healthcheck_success_threshold = 1
}
```

---

## Input Variables

### Provider / Authentication

| Variable | Type | Default | Required | Description |
|----------|------|---------|----------|-------------|
| `kconfig` | `string` | — | **Yes** | Path to the Harvester kubeconfig file. Marked `sensitive = true`; do not commit this value. |
| `kcontext` | `string` | `"local"` | No | Kubeconfig context to use. The default `"local"` is the context name Harvester generates for its own embedded cluster. |

---

### VM Image

These variables control whether the module downloads a fresh image or references one that already exists in Harvester.

| Variable | Type | Default | Required | Description |
|----------|------|---------|----------|-------------|
| `download_image` | `bool` | `false` | No | Set to `true` to create a new `harvester_image` resource from a URL. Set to `false` to reference an existing image via `image_name` / `image_namespace`. |
| `image_name` | `string` | `""` | Conditional | Name of an **existing** VM image in Harvester. Required when `download_image = false`. |
| `image_namespace` | `string` | `""` | Conditional | Namespace of the existing VM image. Required when `download_image = false`. |
| `image_storageclass` | `string` | `""` | Conditional | StorageClass of the existing VM image (e.g., `harvester-longhorn`). Required when `download_image = false`. |
| `new_image` | `object` | Ubuntu 24.04 LTS | Conditional | Configuration for the image to download when `download_image = true`. See schema below. |

#### `new_image` object schema

| Field | Type | Description |
|-------|------|-------------|
| `name` | `string` | Internal resource name (e.g., `"ubuntu24"`). |
| `display_name` | `string` | Human-readable image name shown in the Harvester UI. |
| `source_type` | `string` | Image source type. Currently only `"download"` is supported by the provider. |
| `url` | `string` | HTTP/HTTPS URL of the cloud image (e.g., a `.img` or `.qcow2` file). |

Default value:
```hcl
new_image = {
  name         = "ubuntu24"
  display_name = "noble-server-cloudimg-amd64.img"
  source_type  = "download"
  url          = "https://cloud-images.ubuntu.com/noble/current/noble-server-cloudimg-amd64.img"
}
```

---

### Core VM Configuration

| Variable | Type | Default | Required | Description |
|----------|------|---------|----------|-------------|
| `namespace` | `string` | — | **Yes** | Harvester namespace in which all resources (VMs, secrets, LB) will be created. |
| `vm_count` | `number` | — | **Yes** | Number of identical VMs to create. Each VM is named `<vm_prefix>-<index>` (e.g., `k8s-node-0`, `k8s-node-1`). |
| `vm_prefix` | `string` | — | **Yes** | Name prefix applied to all VMs and derived resources (cloud-init secret, LoadBalancer). |
| `vm_cpu` | `string` | `4` | No | Number of vCPU cores assigned to each VM. |
| `vm_memory` | `string` | `"4Gi"` | No | Memory allocated per VM in Kubernetes quantity format (e.g., `"4Gi"`, `"16Gi"`). |
| `vm_disksize` | `string` | `"40Gi"` | No | Root disk size per VM in Kubernetes quantity format (e.g., `"40Gi"`, `"100Gi"`). |
| `vm_disk_auto_delete` | `bool` | `true` | No | When `true`, the root disk PVC is deleted when the VM is destroyed. Set to `false` to retain disks for data recovery scenarios. |
| `vm_labels` | `map(string)` | `null` | No | Key-value labels applied to each VM resource. Useful for workload identification and selection. Example: `{ env = "staging", role = "worker" }`. |
| `vm_description` | `string` | `""` | No | Free-text description stored on each VM resource (visible in the Harvester UI). |

---

### Networking

| Variable | Type | Default | Required | Description |
|----------|------|---------|----------|-------------|
| `harvester_net` | `string` | — | **Yes** | Name of the Harvester network (VM network / VLAN) to attach to each VM. |
| `harvester_net_namespace` | `string` | — | **Yes** | Namespace in which the Harvester network resource lives. Often the same as `namespace`. |

---

### Cloud-init

The module creates a single `harvester_cloudinit_secret` named `cloud-config-<vm_prefix>` that is shared across all VMs in the group.

| Variable | Type | Default | Required | Description |
|----------|------|---------|----------|-------------|
| `ssh_user` | `string` | `""` | No | SSH login username stored as the `ssh-user` tag on each VM. Used for inventory tooling and documentation purposes; does not directly configure the OS user (use `cloud_config_user_data` for that). |
| `cloud_config_user_data` | `string` | `""` | No | Full cloud-init `user-data` document (typically a `#cloud-config` YAML block). Injected via a Kubernetes secret. See the `terraform.tfvars.example` for a starter template that installs `qemu-guest-agent`. |
| `cloud_config_network_data` | `string` | `""` | No | Cloud-init `network-data` document for advanced network configuration (e.g., static IPs, bonding). Leave empty to use DHCP. |

> **Tip:** The `cloud_config_user_data` example in `terraform.tfvars.example` enables and starts `qemu-guest-agent`, which is required for the post-provisioning IP-wait loop in `main.tf` to work correctly.

---

### LoadBalancer

Setting `create_lb = true` creates a `harvester_loadbalancer` resource named `<vm_prefix>-lb` that targets all VMs in the group via the `terraform-harvester/project: <vm_prefix>` backend selector.

| Variable | Type | Default | Required | Description |
|----------|------|---------|----------|-------------|
| `create_lb` | `bool` | `false` | No | Set to `true` to provision a Harvester LoadBalancer in front of the VM group. |
| `lb_ipam` | `string` | `"dhcp"` | No | IP address allocation method for the LoadBalancer. Valid values: `"dhcp"` or `"pool"` (uses a pre-configured IP pool in Harvester). |
| `lb_listener_port` | `number` | `443` | No | Port on which the LoadBalancer accepts incoming traffic. |
| `lb_listener_backend_port` | `number` | `443` | No | Port on the backend VMs to which traffic is forwarded. |
| `lb_protocol` | `string` | `"TCP"` | No | Transport protocol for the listener. Typically `"TCP"`. |
| `lb_healthcheck_period_seconds` | `number` | `5` | No | Interval in seconds between health check probes. |
| `lb_healthcheck_timeout_seconds` | `number` | `3` | No | Maximum seconds to wait for a health check response before marking it as failed. |
| `lb_healthcheck_failure_threshold` | `number` | `3` | No | Number of consecutive failed probes before a backend is marked unhealthy and removed from rotation. |
| `lb_healthcheck_success_threshold` | `number` | `1` | No | Number of consecutive successful probes before a previously unhealthy backend is returned to rotation. |

---

## Outputs

| Output | Type | Description |
|--------|------|-------------|
| `vm_ip_addresses` | `list(string)` | List of strings in the format `"<vm-name>: <ip1>, <ip2>"` for each provisioned VM. Populated after the `local-exec` IP-wait probe completes. |
| `vm_lb_ip_address` | `string` | The IP address assigned to the Harvester LoadBalancer. Only populated when `create_lb = true |

---

## Resources Created

| Resource | Count | Description |
|----------|-------|-------------|
| `harvester_cloudinit_secret` | 1 | Kubernetes secret holding cloud-init user-data and network-data. |
| `harvester_image` | 0 or 1 | Created only when `download_image = true`. |
| `harvester_virtualmachine` | `vm_count` | Individual VMs named `<vm_prefix>-<index>`. |
| `harvester_loadbalancer` | 0 or 1 | Created only when `create_lb = true`. |

---

## Getting Started

1. **Export your Harvester kubeconfig:**
   ```bash
   # From the Harvester UI: Support → Download KubeConfig
   export KUBECONFIG=/path/to/harvester.yaml
   ```

2. **Copy and edit the example vars file:**
   ```bash
   cp terraform.tfvars.example terraform.tfvars
   # Fill in kconfig, namespace, vm_prefix, harvester_net, etc.
   ```

3. **Initialize and apply:**
   ```bash
   terraform init
   terraform plan
   terraform apply
   ```

4. **Retrieve VM IPs after provisioning:**
   ```bash
   terraform output vm_ip_addresses
   ```

---

## Important Notes

- **`kconfig` is sensitive.** Never commit `terraform.tfvars` containing a kubeconfig path with embedded credentials. Use `TF_VAR_kconfig` or a secrets manager in CI pipelines.
- **IP-wait provisioner requires `kubectl` in `$PATH`** on the machine running `terraform apply`. It polls up to 30 times (5-minute timeout) for the VMI to receive an IP via the guest agent.
- **EFI is always enabled** (`efi = true`, `secure_boot = false`). Ensure your chosen image supports EFI boot (all modern Ubuntu cloud images do).
- **LoadBalancer backend selection** uses the tag `terraform-harvester/project: <vm_prefix>` automatically applied to each VM. Do not remove this tag manually.
- The `vm_lb_ip_address` output will fail if `create_lb = false`. This is a known limitation — wrap it in `try()` if conditionally using the output.

---

## References

- [Harvester HCI Documentation](https://docs.harvesterhci.io/)
- [Harvester Terraform Provider](https://registry.terraform.io/providers/harvester/harvester/latest/docs)
- [Harvester VM Networks](https://docs.harvesterhci.io/en/latest/networking/harvester-network/)
- [cloud-init Documentation](https://cloudinit.readthedocs.io/en/latest/)