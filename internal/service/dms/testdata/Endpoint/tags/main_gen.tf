# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_dms_endpoint" "test" {
  database_name = "tf-test-dms-db"
  endpoint_id   = var.rName
  endpoint_type = "source"
  engine_name   = "aurora"
  password      = "tftest"
  port          = 3306
  server_name   = "tftest"
  ssl_mode      = "none"
  username      = "tftest"

  tags = var.resource_tags
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "resource_tags" {
  description = "Tags to set on resource. To specify no tags, set to `null`"
  # Not setting a default, so that this must explicitly be set to `null` to specify no tags
  type     = map(string)
  nullable = true
}
