---
title: First look at Terraform
slug: first-look-hashicorp-terraform
author: mark
published: 2018-08-05T15:47:00+07:00
tags:
  - terraform
  - hashicorp
  - ops
---

This is a first look at [Terraform](https://www.terraform.io/) from HashiCorp. The most up-to-date work can be found [at this link](https://github.com/exklamationmark/terraform).
# What is it

This is probably an oversimplification of what Terraform is, but let's start anw:

Terraform uses [HashiCorp configuration language](https://github.com/hashicorp/hcl) to describe a **desired** state of infrastructure (e.g: want 2 VM on Google Clould, VM 1 is a `f1-micro`, booting with the image named `ubuntu-18084-bionic-v20180723`, VM2 is a `g1-small`, etc).

After specifying the desired state, we tell Terraform to generate a plan to change our current infrastructure to match it. The plan involves making API calls to the cloud provider (GCP, AWS, Azure, etc). To do this, Terraform uses the current state of infra and make a "patch" to make curren state similar to desired state.

Along the way, the changes are tracked and stored, providing a history. This is stored in [terraform backends](https://www.terraform.io/docs/backends/), which can either be a local file or hosted storage.

The benefits possibly include:

- Managing infrastructure using specification (i.e functional) -> higher level, might be cleaner than keep scripts & procedures to generate them (i.e imperative).
- Abstraction layer to move across cloud providers.
- Document (as code) the state & change history of the infrastructure.

# A very simple example

## Setup (local machine)

- [Download terraform](https://www.terraform.io/downloads.html)
- [Download gclould SDK](https://cloud.google.com/sdk/docs/quickstart-debian-ubuntu)

## Setup (GCP)

- Create a new project
- Create a service account:
	- GCP Console -> IAM & admin -> Service account
	- + Create service account
	- Use Role == `Project -> Owner` for this experiment (full access)
	- Create -> Download the credentials file (e.g: `gcp_creds.json`)
	- Save the secret file somewhere (but don't check in)

## Create a VM

Configs are stored in `examples`, consisting of:

**provider.tf**, defining how to connect to Google API:

<pre class="language-terraform"><code class="language-terraform">
provider "google" {
	credentials = "${file("/path/to/google/credentials/file.json")}"
	project     = "gcp/project/name"
	region      = "asia-southeast1"
}
</code></pre>

**test_vm.tf**, defining a sample VM:

<pre class="language-terraform"><code class="language-terraform">
resource "google_compute_instance" "some-kind-of-name" {
	name         = "VM-name"
	machine_type = "machine-type"
	zone         = "zone-name"

	tags = ["test", "terraform"]

	boot_disk {
		initialize_params {
			image = "image-name"
		}
	}

	network_interface {
		network = "default"
	}
}
</code></pre>

## Operate

<pre class="language-bash" data-user="me" data-host="local"><code class="language-bash">
$ terraform plan examples/            # view the plan (dry-run)
$ terraform apply examples/           # generate a plan && apply
$ terraform destroy examples/         # destroy the setup
</code></pre>

# Cheatsheet for GCP

- Zone name
	- Console -> Compute Engine -> Zones
	- Will have regions like: `asia-southeast1`, `northamerica-northeast1`
	- Each region have 3 zones (e.g: `a, b, c`)
	- Name = `<region-name>-<zone-name>`, e.g: `asia-southeast1-a`, `northamerica-northeast1-b`
- Image name
	- Console -> Compute Engine -> Images
	- Just copy the image name, e.g: `ubuntu-1804-bionic-v20180723`

