output "vm_ip_addresses" {
  value = [for node in resource.harvester_virtualmachine.vm : "${node.name}: ${join(", ", node.network_interface[*].ip_address)}"]
}