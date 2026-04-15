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

data "harvester_virtualmachine" "vm" {
  count      = var.vm_count
  depends_on = [harvester_virtualmachine.vm]
  name       = resource.harvester_virtualmachine.vm[count.index].name
  namespace  = var.namespace
}

data "harvester_loadbalancer" "vm_lb" {
  count = resource.harvester_loadbalancer.vm_lb.count
  name = resource.harvester_loadbalancer.vm_lb.name
  namespace = var.namespace
  depends_on = [ harvester_loadbalancer.vm_lb ]
}