package mysql_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/NaverCloudPlatform/ncloud-sdk-go-v2/services/vmysql"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	. "github.com/terraform-providers/terraform-provider-ncloud/internal/acctest"
	"github.com/terraform-providers/terraform-provider-ncloud/internal/conn"
	mysqlservice "github.com/terraform-providers/terraform-provider-ncloud/internal/service/mysql"
)

func TestAccResourceNcloudMysqlRecovery_vpc_basic(t *testing.T) {
	var mysqlSErverInstance vmysql.CloudMysqlServerInstance
	testName := fmt.Sprintf("tf-mysqlsv-%s", acctest.RandString(5))
	resourceName := "ncloud_mysql_recovery.mysql_recovery"
	testDate := "20240920"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { TestAccPreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckMysqlRecoveryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMysqlRecoveryConfig(testName, testDate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMysqlRecoveryExists(resourceName, &mysqlSErverInstance, GetTestProvider(true)),
					resource.TestCheckResourceAttrSet(resourceName, "mysql_instance_no"),
				),
			},
		},
	})
}

func testAccMysqlRecoveryConfig(testName string, testDate string) string {
	return fmt.Sprintf(`
data "ncloud_vpc" "test_vpc" {
	id = "75658"
}
data "ncloud_subnet" "test_subnet" {
	id = "172709"
}

resource "ncloud_mysql" "mysql" {
	subnet_no = data.ncloud_subnet.test_subnet.id
	service_name = "%[1]s"
	server_name_prefix = "testprefix"
	user_name = "testusername"
	user_password = "t123456789!a"
	host_ip = "192.168.0.1"
	database_name = "test_db"
}

resource "ncloud_mysql_recovery" "mysql_recovery" {
	mysql_instance_no = ncloud_mysql.mysql.id
	recovery_server_name = "test-recovery"
	file_name = "%[2]s"
}
`, testName, testDate)
}

func testAccCheckMysqlRecoveryExists(n string, recovery *vmysql.CloudMysqlServerInstance, provider *schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resource, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found %s", n)
		}

		if resource.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		config := provider.Meta().(*conn.ProviderConfig)
		mysqlRecovery, err := mysqlservice.GetMysqlRecovery(context.Background(), config, resource.Primary.Attributes["mysql_instance_no"])
		if err != nil {
			return nil
		}

		if mysqlRecovery != nil {
			*recovery = *mysqlRecovery
			return nil
		}

		return fmt.Errorf("mysql recovery not found")
	}
}

func testAccCheckMysqlRecoveryDestroy(s *terraform.State) error {
	config := GetTestProvider(true).Meta().(*conn.ProviderConfig)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ncloud_mysql_recovery" {
			continue
		}
		instance, err := mysqlservice.GetMysqlRecovery(context.Background(), config, rs.Primary.Attributes["mysql_instance_no"])
		if err != nil && !strings.Contains(err.Error(), "5001017") {
			return nil
		}

		if instance != nil {
			return errors.New("mysql recovery still exists")
		}
	}

	return nil
}
