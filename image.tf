resource "harvester_image" "new_image" {
  count              = var.download_image ? 1 : 0
  name               = var.new_image.name
  namespace          = var.image_namespace
  display_name       = var.new_image.display_name
  source_type        = var.new_image.source_type
  url                = var.new_image.url
  storage_class_name = data.harvester_storageclass.longhorn.name
}
