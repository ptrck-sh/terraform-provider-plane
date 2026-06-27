terraform {
  required_providers {
    plane = {
      source = "ptrck-sh/plane"
    }
  }
}

provider "plane" {
  host    = "https://plane.example.com" # or PLANE_HOST
  api_key = var.plane_api_key           # or PLANE_API_KEY
}

variable "plane_api_key" {
  type      = string
  sensitive = true
}
