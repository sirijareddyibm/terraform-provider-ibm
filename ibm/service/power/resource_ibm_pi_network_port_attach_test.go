// Copyright IBM Corp. 2017, 2021 All Rights Reserved.
// Licensed under the Mozilla Public License v2.0

package power_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	acc "github.com/IBM-Cloud/terraform-provider-ibm/ibm/acctest"

	"github.com/IBM-Cloud/power-go-client/clients/instance"
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/conns"
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/flex"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccIBMPINetworkPortAttachbasic(t *testing.T) {
	name := fmt.Sprintf("tf-pi-instance-%d", acctest.RandIntRange(10, 100))
	networkName := fmt.Sprintf("tf-pi-network-port-attach-test-%d", acctest.RandIntRange(10, 100))
	networkName2 := fmt.Sprintf("tf-pi-network-port-attach-test-%d", acctest.RandIntRange(10, 100))
	health := "OK"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acc.TestAccPreCheck(t) },
		Providers:    acc.TestAccProviders,
		CheckDestroy: testAccCheckIBMPINetworkPortAttachDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckIBMPINetworkPortAttachConfig(name, networkName, networkName2, health),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIBMPINetworkPortAttachExists("ibm_pi_network_port_attach.power_network_port_attach"),
					resource.TestCheckResourceAttr(
						"ibm_pi_network_port_attach.power_network_port_attach", "pi_network_name", networkName2),
					resource.TestCheckResourceAttrSet("ibm_pi_network_port_attach.power_network_port_attach", "id"),
					resource.TestCheckResourceAttrSet("ibm_pi_network_port_attach.power_network_port_attach", "network_port_id"),
					resource.TestCheckResourceAttrSet("ibm_pi_network_port_attach.power_network_port_attach", "public_ip"),
				),
			},
		},
	})
}

func TestAccIBMPINetworkPortAttachVlanbasic(t *testing.T) {
	name := fmt.Sprintf("tf-pi-instance-%d", acctest.RandIntRange(10, 100))
	networkName := fmt.Sprintf("tf-pi-network-port-attach-test-%d", acctest.RandIntRange(10, 100))
	networkName2 := fmt.Sprintf("tf-pi-network-port-attach-test-%d", acctest.RandIntRange(10, 100))
	health := "OK"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acc.TestAccPreCheck(t) },
		Providers:    acc.TestAccProviders,
		CheckDestroy: testAccCheckIBMPINetworkPortAttachDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckIBMPINetworkPortAttachVlanConfig(name, networkName, networkName2, health),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIBMPINetworkPortAttachExists("ibm_pi_network_port_attach.power_network_port_attach"),
					resource.TestCheckResourceAttr(
						"ibm_pi_network_port_attach.power_network_port_attach", "pi_network_name", networkName2),
					resource.TestCheckResourceAttrSet("ibm_pi_network_port_attach.power_network_port_attach", "id"),
					resource.TestCheckResourceAttrSet("ibm_pi_network_port_attach.power_network_port_attach", "network_port_id"),
				),
			},
		},
	})
}

func testAccCheckIBMPINetworkPortAttachDestroy(s *terraform.State) error {
	sess, err := acc.TestAccProvider.Meta().(conns.ClientSession).IBMPISession()
	if err != nil {
		return err
	}
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ibm_pi_network_port_attach" {
			continue
		}
		parts, err := flex.IdParts(rs.Primary.ID)
		if err != nil {
			return err
		}
		cloudInstanceID := parts[0]
		networkname := parts[1]
		portID := parts[2]
		networkC := instance.NewIBMPINetworkClient(context.Background(), sess, cloudInstanceID)
		_, err = networkC.GetPort(networkname, portID)
		if err == nil {
			return fmt.Errorf("PI Network Port still exists: %s", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckIBMPINetworkPortAttachExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No Record ID is set")
		}

		sess, err := acc.TestAccProvider.Meta().(conns.ClientSession).IBMPISession()
		if err != nil {
			return err
		}
		parts, err := flex.IdParts(rs.Primary.ID)
		if err != nil {
			return err
		}
		cloudInstanceID := parts[0]
		networkname := parts[1]
		portID := parts[2]
		client := instance.NewIBMPINetworkClient(context.Background(), sess, cloudInstanceID)

		_, err = client.GetPort(networkname, portID)
		if err != nil {
			return err
		}
		return nil
	}
}

func testAccCheckIBMPINetworkPortAttachConfig(name, networkName, networkName2, health string) string {
	return fmt.Sprintf(`
		data "ibm_pi_image" "power_image" {
			pi_cloud_instance_id = "%[1]s"
			pi_image_name        = "%[3]s"
		}
		resource "ibm_pi_network" "power_networks" {
			pi_cloud_instance_id = "%[1]s"
			pi_cidr              = "192.168.15.0/24"
			pi_network_name      = "%[5]s"
			pi_network_type      = "vlan"
		}
		resource "ibm_pi_network" "power_networks2" {
			pi_cloud_instance_id = "%[1]s"
			pi_network_name      = "%[6]s"
			pi_network_type      = "pub-vlan"
		}
		resource "ibm_pi_instance" "power_instance" {
			pi_cloud_instance_id  = "%[1]s"
			pi_image_id           = data.ibm_pi_image.power_image.id
			pi_instance_name      = "%[2]s"
			pi_memory             = "2"
			pi_proc_type          = "shared"
			pi_processors         = "0.25"
			pi_storage_pool       = data.ibm_pi_image.power_image.storage_pool
			pi_storage_type       = "%[4]s"
			pi_health_status      = "%[7]s"
			pi_sys_type           = "s922"
			pi_network {
				network_id = resource.ibm_pi_network.power_networks.network_id
			}
		}
		resource "ibm_pi_network_port_attach" "power_network_port_attach" {
			pi_cloud_instance_id        = "%[1]s"
			pi_instance_id              = resource.ibm_pi_instance.power_instance.instance_id
			pi_network_name             = resource.ibm_pi_network.power_networks2.pi_network_name 
			pi_network_port_description = "IP Reserved for Test UAT"
		}
	`, acc.Pi_cloud_instance_id, name, acc.Pi_image, acc.PiStorageType, networkName, networkName2, health)
}

func testAccCheckIBMPINetworkPortAttachVlanConfig(name, networkName, networkName2, health string) string {
	return fmt.Sprintf(`
		data "ibm_pi_image" "power_image" {
			pi_cloud_instance_id = "%[1]s"
			pi_image_name        = "%[3]s"
		}
		resource "ibm_pi_network" "power_networks" {
			pi_cloud_instance_id = "%[1]s"
			pi_cidr              = "192.168.15.0/24"
			pi_network_name      = "%[5]s"
			pi_network_type      = "vlan"
		}
		resource "ibm_pi_network" "power_networks2" {
			pi_cloud_instance_id = "%[1]s"
			pi_cidr              = "192.97.57.0/24"
			pi_network_name      = "%[6]s"
			pi_network_type      = "vlan"
		}
		resource "ibm_pi_instance" "power_instance" {
			pi_cloud_instance_id  = "%[1]s"
			pi_image_id           = data.ibm_pi_image.power_image.id
			pi_instance_name      = "%[2]s"
			pi_memory             = "2"
			pi_proc_type          = "shared"
			pi_processors         = "0.25"
			pi_storage_pool       = data.ibm_pi_image.power_image.storage_pool
			pi_storage_type       = "%[4]s"
			pi_health_status      = "%[7]s"
			pi_sys_type           = "s922"
			pi_network {
				network_id = resource.ibm_pi_network.power_networks.network_id
			}
		}
		resource "ibm_pi_network_port_attach" "power_network_port_attach" {
			pi_cloud_instance_id        = "%[1]s"
			pi_instance_id              = resource.ibm_pi_instance.power_instance.instance_id
			pi_network_name             = ibm_pi_network.power_networks2.pi_network_name
			pi_network_port_description = "IP Reserved for Test UAT"
		}
	`, acc.Pi_cloud_instance_id, name, acc.Pi_image, acc.PiStorageType, networkName, networkName2, health)
}
