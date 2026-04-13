data "harvester_image" "image" {
  display_name = var.download_image ? resource.harvester_image.new_image[0].display_name : var.image_name
  namespace    = var.image_namespace
}

data "harvester_network" "vm_net" {
  name      = var.harvester_net
  namespace = var.harvester_net_namespace
}

data "harvester_storageclass" "longhorn" {
  name = var.image_storageclass
}