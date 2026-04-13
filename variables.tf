variable "kconfig" {
  type        = string
  description = "Path to Harvester kubeconfig"
  sensitive   = true
}

variable "kcontext" {
  type        = string
  description = "Specify Harvester kubeconfig context"
  default     = "local"
}

variable "download_image" {
  type        = bool
  default     = false
  description = "Create a new harvester image resource"
}

variable "image_name" {
  type        = string
  description = "Name of existing vm image"
  default     = ""
}

variable "image_namespace" {
  type        = string
  description = "Namespace containing existing vm image"
  default     = ""
}

variable "image_storageclass" {
  type        = string
  description = "StorageClass of existing vm image"
  default     = ""
}

variable "new_image" {
  type = object({
    name         = string
    display_name = string
    source_type  = string
    url          = string
  })

  default = {
    name         = "ubuntu24"
    display_name = "noble-server-cloudimg-amd64.img"
    source_type  = "download"
    url          = "https://cloud-images.ubuntu.com/noble/current/noble-server-cloudimg-amd64.img"
  }
}

variable "namespace" {
  type        = string
  description = "Harvester resource namespace"
}

variable "vm_count" {
  type        = number
  description = "Count of identical VMs to provision"
}

variable "vm_prefix" {
  type        = string
  description = "VM name prefix"
}

variable "vm_description" {
  type        = string
  description = "Description for provisioned VMs"
  default     = ""
}

variable "vm_cpu" {
  type        = string
  description = "VM CPU core count"
  default     = 4
}

variable "vm_memory" {
  type        = string
  description = "VM memory amount"
  default     = "4Gi"
}

variable "vm_disksize" {
  type        = string
  description = "VM rootdisk size"
  default     = "40Gi"
}

variable "vm_disk_auto_delete" {
  type        = bool
  description = "Set to false to retain VM disks after VM deletion"
  default     = true
}

variable "harvester_net" {
  type        = string
  description = "Harvester network name"
}

variable "harvester_net_namespace" {
  type        = string
  description = "Harvester network namespace"
}

variable "ssh_user" {
  type        = string
  description = "Account used to connect to vms"
  default     = ""
}

variable "cloud_config_user_data" {
  type        = string
  description = "cloud-init user-data"
  default     = ""
}