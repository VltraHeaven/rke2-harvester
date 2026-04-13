resource "harvester_cloudinit_secret" "cloud-config" {
  name         = "cloud-config-${var.vm_prefix}"
  namespace    = var.namespace
  user_data    = var.cloud_config_user_data
  network_data = ""
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

  cpu         = 4
  memory      = "4Gi"
  efi         = true
  secure_boot = false

  network_interface {
    name         = "nic-1"
    network_name = data.harvester_network.vm_net.id
  }


  disk {
    name        = "rootdisk"
    type        = "disk"
    size        = "40Gi"
    bus         = "virtio"
    boot_order  = 1
    image       = data.harvester_image.image.id
    auto_delete = true
  }

  input {
    name = "tablet"
    type = "tablet"
    bus  = "usb"
  }
  depends_on = [ data.harvester_image.image ]
}