variable "TAG" {
  default = "local"
}
variable "REPO" {
  default = "ghcr.io/lesomnus/tegra-exporter"
}
variable "BUILD_HASH" {
  default = "0000000000000000000000000000000000000000"
}
variable "BUILD_DATE" {
  default = "${formatdate("YYMMDD", timestamp())}"
}
variable "BUILD_ID" {
  default = "r0"
}

target "test" {
  target = "test"
}
target "app" {
  args = {
    BUILD_HASH = BUILD_HASH
    BUILD_DATE = BUILD_DATE
    BUILD_ID   = BUILD_ID
  }
  tags = [
    "${REPO}:${TAG}",
    "${REPO}:${BUILD_ID}",
    "${REPO}:${BUILD_DATE}",
    "${REPO}:${BUILD_DATE}-${BUILD_ID}",
    "${REPO}:${BUILD_HASH}",
  ]
  target = "app"
}
