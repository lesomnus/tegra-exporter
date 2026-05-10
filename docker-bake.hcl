variable "TAG" {
  default = "local"
}
variable "REPO" {
  default = "ghcr.io/lesomnus/tegra-exporter"
}
variable "BUILD_HASH" {
  default = "0000000000000000000000000000000000000000"
}
variable "BUILD_ID" {
  default = "r0"
}

target "app" {
  tags = [
    "${REPO}:${TAG}",
    "${REPO}:${BUILD_ID}",
  ]
}
