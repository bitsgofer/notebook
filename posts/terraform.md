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

This is a first look at [Terraform](https://www.terraform.io/) from HashiCorp.

# What is Terraform?

This is probably an oversimplification of what Terraform is, but let's start anw:

Terraform uses [HashiCorp configuration language](https://github.com/hashicorp/hcl)
to describe a **desired** state of infrastructure (e.g: want 2 VM on Google Clould,
VM 1 is a `f1-micro`, booting with the image named `ubuntu-18084-bionic-v20180723`,
VM2 is a `g1-small`, etc).

After specifying the desired state, we tell Terraform to generate a plan
to change our current infrastructure to match it.
The plan involves making API calls to the cloud provider (GCP, AWS, Azure, etc).
To do this, Terraform uses the current state of infra and make a "patch"
to make current state similar to desired state.

Along the way, the changes are tracked and stored, providing a history.
This is stored in [terraform backends](https://www.terraform.io/docs/backends/),
which can either be a local file or hosted storage.

The benefits possibly include:

- Managing infrastructure using specification (i.e functional) -> higher level,
  might be cleaner than keep scripts & procedures to generate them (i.e imperative).
- ~~Abstraction layer to move across cloud providers.~~
  You have to things specifically for your clould provider.
- Document (as code) the current state (present \*.tf files) & change history
  (Terraform state + changes on \*.tf files) of the infrastructure.


# Using Terraform

I tried to use Terraform to setup a simple project on Google Clould:

- Deploy a VPC
- Deploy a VM to use as SSH jump host, using a custom port

After playing around, I ended up creating the desired setup (see the code section below).

My take from this is that Terraform, while useful in its own way, is not a be-all and end-all tool
that I first perceived it to be. In particular:

## Strength

Assuming your \*.tf files are up-to-date, reading through all these files tell you about the
current state of your infrastructure (how many VMs are there, what are the VPCs, subnets, etc).

You don't need to click a lot of buttons cloud provider's interface & try to document what to click.
**Most** common resources can be specified in a \*.tf files. Terraform then make the API calls
to created/update them.

When you need to change your infrastructure's configuration, sometimes Terraform can figure out
whether you can make an edit on an existing resource or re-create them from scratch.
This means you can focus on the end-goal and (usually) not worry about the steps in between.

All in all, it makes for a very good experience when you need do to provision some infrastructure,
especially if you are doing it from scratch.

## Weakness

You have to trade off precise control for declarative configuration. Because you can't really
tell Terraform a list of things to execute, things get akward when you need to change a live system.
For example, because I only have 1 SSH jump host, changes that requires re-creating the VM
will terminate all SSH sessions currently on it.

Unless all your stuff can be modified in-place or run in a cluster, thus can tolerate
total VM failures, using Terraform later on might not be ideal.

Plus, you have to be careful when running `terraform apply`. It might bring down something like your
only DB VM. Protection might be required so that you don't accidentally hurt your self.

Another quirk is that because Terraform is only supposed to keep the current state in place,
certain tasks might not be easily described this way. For example, when I need to setup the
SSH jump host to listen on a custom port, I need to:

- Deploy with a firewall rule for port 22 (because sshd listens on this by default)
- Declare a `remote-exec` block that changes the port and then restart `ssh.service`
- Update the firewall rule to change the allow port from 22 to something else

-> not purely declarative anymore


## Thoughts

Terraform might be the first tool I reach out for when I need to create some resources. Hopefull
it can replaces the bunch of bash scripts that I really hate to write.

However, after provisioning them, I might lean to something else (Ansible / more bash scripts)
to change things the right way.

Another interesting bits is the possibility to use Terraform to work with stuff like Kubernetes,
since there's some similarities between both (and Helm, perhaps).

## Rabbit holes

- To what degree should we rely on the default values that Terraform provides?
  It can be useful (shorter files), but also hide things.
- How do you test that your Terraform files works?
  Will you be forced to waste money on a clould provider?
- How does Terraform knows when a resource should be modified in-place vs when it should be created new?
- Changing some block doesn't trigger update when calling `apply` (e.g: `remote-exec` provisioner).
  You have to taint the resource manually to let Terraform knows. Can this be detected?

# Code

### provider.tf

<pre class="language-hcl"><code class;s"language-hcl">
provider "google" {
	credentials = "${file("terraform-credentials.json")}"
	project     = "my-project-id"
    region      = "asia-southeast1"
}
</code></pre>


### vpc.tf
<pre class="language-hcl"><code class;s"language-hcl">
resource "google_compute_network" "projectName" {
	name = "projectName"
	auto_create_subnetworks = true
}

resource "google_compute_firewall" "projectName-allow-icmp" {
	name    = "projectName-allow-icmp"
	network = "${google_compute_network.projectName.name}"

	disabled = false
	direction = "INGRESS"
	allow {
		protocol = "icmp"
	}
	source_ranges = ["0.0.0.0/0"]
	priority = "65534"
}

resource "google_compute_firewall" "projectName-allow-internal" {
	name    = "projectName-allow-internal"
	network = "${google_compute_network.projectName.name}"

	disabled = false
	direction = "INGRESS"
	allow {
		protocol = "all"
	}
	source_ranges = ["10.128.0.0/9"]
	priority = "65534"
}

resource "google_compute_firewall" "projectName-allow-ssh-22" {
	name    = "projectName-allow-ssh-22"
	network = "${google_compute_network.projectName.name}"

	disabled = false
	direction = "INGRESS"
	allow {
		protocol = "tcp"
		ports = ["22"]
	}
	source_ranges = ["0.0.0.0/0"]
	priority = "65534"
}

resource "google_compute_firewall" "projectName-allow-ssh-2792" {
	name    = "projectName-allow-ssh-2792"
	network = "${google_compute_network.projectName.name}"

	disabled = false
	direction = "INGRESS"
	allow {
		protocol = "tcp"
		ports = ["2792"]
	}
	source_ranges = ["0.0.0.0/0"]
	priority = "65534"
}

resource "google_compute_firewall" "projectName-deny-all-ingress" {
	name    = "projectName-deny-all-ingress"
	network = "${google_compute_network.projectName.name}"

	disabled = false
	direction = "INGRESS"
	allow {
		protocol = "all"
	}
	source_ranges = ["0.0.0.0/0"]
	priority = "65535"
}

resource "google_compute_firewall" "projectName-deny-all-egress" {
	name    = "projectName-deny-all-egress"
	network = "${google_compute_network.projectName.name}"

	disabled = false
	direction = "EGRESS"
	allow {
		protocol = "all"
	}
	destination_ranges = ["0.0.0.0/0"]
	priority = "65535"
}
</code></pre>

### server.tf

<pre class="language-hcl"><code class;s"language-hcl">
resource "google_compute_address" "eden_public_ip" {
	name = "projectName-public-ip"
	address_type = "EXTERNAL"
	network_tier = "PREMIUM"
}

resource "google_compute_instance" "projectName" {
	name = "projectName"
    description = "home server, hosting sshd, blog sever && vault"
	tags = ["home", "ssh-jumphost", "blog-server", "vault-server"]

	deletion_protection = false # TODO(mark): set to true when done
	machine_type				= "f1-micro"
	allow_stopping_for_update	= true
	zone 						= "asia-southeast1-a"

	boot_disk {
		initialize_params {
			image = "ubuntu-1804-bionic-v20180814"
		}
	}

	network_interface {
		network = "${google_compute_network.projectName.name}"
		access_config {
			network_tier = "PREMIUM"
			nat_ip = "${google_compute_address.eden_public_ip.address}"
		}
	}

	can_ip_forward = false

	scheduling {
		preemptible = false
		on_host_maintenance = "MIGRATE"
		automatic_restart = true
	}

	provisioner "remote-exec" {
		connection {
			type = "ssh"
			user = "terraform-agent"
			port = "22" ## We will chagne this to 2792 now
			private_key = "${file("../../.ssh/terraform-agent")}"
		}

		inline = [
			"sudo sed -i /etc/ssh/sshd_config -e 's/#Port 22/Port 2792/g'",
			"sudo systemctl restart ssh.service"
		]
	}
}
</code></pre>



# Cheatsheet for GCP

## Setup (local machine)

- [Download terraform](https://www.terraform.io/downloads.html)
- [Download gclould SDK](https://cloud.google.com/sdk/docs/quickstart-debian-ubuntu)

## Useful things

- Zone name
	- Console -> Compute Engine -> Zones
	- Will have regions like: `asia-southeast1`, `northamerica-northeast1`
	- Each region have 3 zones (e.g: `a, b, c`)
	- Name = `<region-name>-<zone-name>`, e.g: `asia-southeast1-a`, `northamerica-northeast1-b`
- Image name
	- Console -> Compute Engine -> Images
	- Just copy the image name, e.g: `ubuntu-1804-bionic-v20180723`

