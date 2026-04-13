resource "harvester_cloudinit_secret" "cloud-config" {
  name         = "cloud-config-${var.vm_prefix}"
  namespace    = var.namespace
  user_data    = var.cloud_config_user_data
  network_data = var.cloud_config_network_data
}

resource "harvester_virtualmachine" "vm" {
  count       = var.vm_count
  name        = "${var.vm_prefix}-${count.index}"
  namespace   = var.namespace
  description = var.vm_description

  tags = {
    ssh-user = var.ssh_user
  }

  cloudinit {
    user_data_secret_name    = harvester_cloudinit_secret.cloud-config.name
    network_data_secret_name = harvester_cloudinit_secret.cloud-config.name
  }

  cpu         = var.vm_cpu
  memory      = var.vm_memory
  efi         = true
  secure_boot = false

  network_interface {
    name         = "nic-1"
    network_name = data.harvester_network.vm_net.id
  }


  disk {
    name        = "rootdisk"
    type        = "disk"
    size        = var.vm_disksize
    bus         = "virtio"
    boot_order  = 1
    image       = data.harvester_image.image.id
    auto_delete = var.vm_disk_auto_delete
  }

  input {
    name = "tablet"
    type = "tablet"
    bus  = "usb"
  }
  depends_on = [data.harvester_image.image]

  provisioner "local-exec" {
  interpreter = ["/bin/bash", "-c"]
  command     = <<-EOT
    for i in $(seq 1 30); do
      IP=$(kubectl get vmi ${var.vm_prefix}-${count.index} -n ${var.namespace} \
        -o jsonpath='{.status.interfaces[0].ipAddress}' 2>/dev/null)
      if [[ -n "$IP" && "$IP" != "<none>" ]]; then
        echo "VM IP: $IP"
        exit 0
      fi
      echo "Waiting for IP... attempt $i"
      sleep 10
    done
    echo "Timed out waiting for IP"
    exit 1
  EOT
  }
}