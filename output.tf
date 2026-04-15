output "vm_ip_addresses" {
  value      = [for vm in data.harvester_virtualmachine.vm : "${vm.name}: ${join(", ", vm.network_interface[*].ip_address)}"]
  depends_on = [harvester_virtualmachine.vm]
}

output "vm_lb_ip_address" {
  value      = data.harvester_loadbalancer.vm_lb[0].ip_address
  depends_on = [data.harvester_loadbalancer.vm_lb]
}