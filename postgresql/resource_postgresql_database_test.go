package postgresql

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccPostgresqlDatabase_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPostgresqlDatabaseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPostgreSQLDatabaseConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPostgresqlDatabaseExists("postgresql_database.mydb", nil),
					resource.TestCheckResourceAttr(
						"postgresql_database.mydb", "name", "mydb"),
					resource.TestCheckResourceAttr(
						"postgresql_database.mydb", "owner", "myrole"),
					resource.TestCheckResourceAttr(
						"postgresql_database.default_opts", "owner", "myrole"),
					resource.TestCheckResourceAttr(
						"postgresql_database.default_opts", "name", "default_opts_name"),
					resource.TestCheckResourceAttr(
						"postgresql_database.default_opts", "template", "template0"),
					resource.TestCheckResourceAttr(
						"postgresql_database.default_opts", "encoding", "UTF8"),
					resource.TestCheckResourceAttr(
						"postgresql_database.default_opts", "lc_collate", "C"),
					resource.TestCheckResourceAttr(
						"postgresql_database.default_opts", "lc_ctype", "C"),
					resource.TestCheckResourceAttr(
						"postgresql_database.default_opts", "tablespace_name", "pg_default"),
					resource.TestCheckResourceAttr(
						"postgresql_database.default_opts", "connection_limit", "-1"),
					resource.TestCheckResourceAttr(
						"postgresql_database.default_opts", "is_template", "false"),

					resource.TestCheckResourceAttr(
						"postgresql_database.modified_opts", "owner", "myrole"),
					resource.TestCheckResourceAttr(
						"postgresql_database.modified_opts", "name", "custom_template_db"),
					resource.TestCheckResourceAttr(
						"postgresql_database.modified_opts", "template", "template0"),
					resource.TestCheckResourceAttr(
						"postgresql_database.modified_opts", "encoding", "UTF8"),
					resource.TestCheckResourceAttr(
						"postgresql_database.modified_opts", "lc_collate", "en_US.UTF-8"),
					resource.TestCheckResourceAttr(
						"postgresql_database.modified_opts", "lc_ctype", "en_US.UTF-8"),
					resource.TestCheckResourceAttr(
						"postgresql_database.modified_opts", "tablespace_name", "pg_default"),
					resource.TestCheckResourceAttr(
						"postgresql_database.modified_opts", "connection_limit", "10"),
					resource.TestCheckResourceAttr(
						"postgresql_database.modified_opts", "is_template", "true"),

					resource.TestCheckResourceAttr(
						"postgresql_database.pathological_opts", "owner", "myrole"),
					resource.TestCheckResourceAttr(
						"postgresql_database.pathological_opts", "name", "bad_template_db"),
					resource.TestCheckResourceAttr(
						"postgresql_database.pathological_opts", "template", "template0"),
					resource.TestCheckResourceAttr(
						"postgresql_database.pathological_opts", "encoding", "LATIN1"),
					resource.TestCheckResourceAttr(
						"postgresql_database.pathological_opts", "lc_collate", "C"),
					resource.TestCheckResourceAttr(
						"postgresql_database.pathological_opts", "lc_ctype", "C"),
					resource.TestCheckResourceAttr(
						"postgresql_database.pathological_opts", "tablespace_name", "pg_default"),
					resource.TestCheckResourceAttr(
						"postgresql_database.pathological_opts", "connection_limit", "0"),
					resource.TestCheckResourceAttr(
						"postgresql_database.pathological_opts", "is_template", "true"),
					resource.TestCheckResourceAttr(
						"postgresql_database.pathological_opts", "search_path.#", "0"),

					resource.TestCheckResourceAttr(
						"postgresql_database.pg_default_opts", "owner", "myrole"),
					resource.TestCheckResourceAttr(
						"postgresql_database.pg_default_opts", "name", "pg_defaults_db"),
					resource.TestCheckResourceAttr(
						"postgresql_database.pg_default_opts", "template", "DEFAULT"),
					// resource.TestCheckResourceAttr(
					// 	"postgresql_database.pg_default_opts", "encoding", "DEFAULT"),
					// resource.TestCheckResourceAttr(
					// 	"postgresql_database.pg_default_opts", "lc_collate", "DEFAULT"),
					// resource.TestCheckResourceAttr(
					//  "postgresql_database.pg_default_opts", "lc_ctype", "DEFAULT"),
					// resource.TestCheckResourceAttr(
					// 	"postgresql_database.pg_default_opts", "tablespace_name", "DEFAULT"),
					resource.TestCheckResourceAttr(
						"postgresql_database.pg_default_opts", "connection_limit", "0"),
					resource.TestCheckResourceAttr(
						"postgresql_database.pg_default_opts", "is_template", "true"),

					testAccCheckPostgresqlDatabaseExists("postgresql_database.mydb_search_path", []string{"bar", "foo-with-hyphen"}),
				),
			},
		},
	})
}

func TestAccPostgresqlDatabase_DefaultOwner(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPostgresqlDatabaseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPostgreSQLDatabaseConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPostgresqlDatabaseExists("postgresql_database.mydb_default_owner", nil),
					resource.TestCheckResourceAttr(
						"postgresql_database.mydb_default_owner", "name", "mydb_default_owner"),
					resource.TestCheckResourceAttrSet(
						"postgresql_database.mydb_default_owner", "owner"),
				),
			},
		},
	})
}

func TestAccPostgresqlDatabase_Update(t *testing.T) {

	// Version dependent features values will be set in PreCheck
	// because we need to access database to check Postgres version.

	// Allow connection depends of Postgres version (needs pg >= 9.5)
	var allowConnections bool

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)

			client := testAccProvider.Meta().(*Client)
			db, err := client.Connect()
			if err != nil {
				t.Fatalf("could not connect to database: %v", err)
			}
			allowConnections = db.featureSupported(featureDBAllowConnections)

		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPostgresqlDatabaseDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource postgresql_database test_db {
    name = "test_db"
	allow_connections = "%t"
    search_path = ["searchpathInitial"]
}
`, allowConnections),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPostgresqlDatabaseExists("postgresql_database.test_db", []string{"searchpathInitial"}),
					resource.TestCheckResourceAttr("postgresql_database.test_db", "name", "test_db"),
					resource.TestCheckResourceAttr("postgresql_database.test_db", "connection_limit", "-1"),
					resource.TestCheckResourceAttr(
						"postgresql_database.test_db", "allow_connections",
						strconv.FormatBool(allowConnections),
					),
				),
			},
			{
				Config: `
resource postgresql_database test_db {
	name = "test_db"
	connection_limit = 2
	allow_connections = false
	search_path = ["searchpathUpdated"]
}
	`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPostgresqlDatabaseExists("postgresql_database.test_db", []string{"searchpathUpdated"}),
					resource.TestCheckResourceAttr("postgresql_database.test_db", "name", "test_db"),
					resource.TestCheckResourceAttr("postgresql_database.test_db", "connection_limit", "2"),
					resource.TestCheckResourceAttr(
						"postgresql_database.test_db", "allow_connections", "false",
					),
				),
			},
		},
	})
}

// Test the case where we need to grant the owner to the connected user.
// The owner should be revoked
func TestAccPostgresqlDatabase_GrantOwner(t *testing.T) {
	skipIfNotAcc(t)

	config := getTestConfig(t)
	dsn := config.connStr("postgres")

	var stateConfig = `
resource postgresql_role "test_owner" {
       name = "test_owner"
}
resource postgresql_database "test_db" {
       name  = "test_db"
       owner = "${postgresql_role.test_owner.name}"
}
`
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPostgresqlDatabaseDestroy,
		Steps: []resource.TestStep{
			{
				Config: stateConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPostgresqlDatabaseExists("postgresql_database.test_db", nil),
					resource.TestCheckResourceAttr("postgresql_database.test_db", "name", "test_db"),
					resource.TestCheckResourceAttr("postgresql_database.test_db", "owner", "test_owner"),

					// check if connected user does not have test_owner granted anymore.
					checkUserMembership(t, dsn, config.Username, "test_owner", false),
				),
			},
		},
	})
}

// Test the case where the connected user is already a member of the owner.
// There were a bug which was revoking the owner anyway.
func TestAccPostgresqlDatabase_GrantOwnerNotNeeded(t *testing.T) {
	skipIfNotAcc(t)

	config := getTestConfig(t)
	dsn := config.connStr("postgres")

	dbExecute(
		t, dsn,
		fmt.Sprintf("CREATE ROLE test_owner; GRANT test_owner TO %s", config.Username),
	)
	defer func() {
		dbExecute(t, dsn, "DROP ROLE test_owner")
	}()

	var stateConfig = `
resource postgresql_database "test_db" {
       name  = "test_db"
       owner = "test_owner"
}
`
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPostgresqlDatabaseDestroy,
		Steps: []resource.TestStep{
			{
				Config: stateConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPostgresqlDatabaseExists("postgresql_database.test_db", nil),
					resource.TestCheckResourceAttr("postgresql_database.test_db", "name", "test_db"),
					resource.TestCheckResourceAttr("postgresql_database.test_db", "owner", "test_owner"),

					// check if connected user still have test_owner granted.
					checkUserMembership(t, dsn, config.Username, "test_owner", true),
				),
			},
		},
	})
}

func checkUserMembership(
	t *testing.T, dsn, member, role string, shouldHaveRole bool,
) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		db, err := sql.Open("postgres", dsn)
		if err != nil {
			t.Fatalf("could to create connection pool: %v", err)
		}
		defer db.Close()

		var _rez int
		err = db.QueryRow(`
                       SELECT 1 FROM pg_auth_members
                       WHERE pg_get_userbyid(roleid) = $1 AND pg_get_userbyid(member) = $2
               `, role, member).Scan(&_rez)

		switch {
		case err == sql.ErrNoRows:
			if shouldHaveRole {
				return fmt.Errorf(
					"User %s is not a member of %s",
					member, role,
				)
			}
			return nil

		case err != nil:
			t.Fatalf("could not check granted role: %v", err)
		}

		if !shouldHaveRole {
			return fmt.Errorf(
				"User (%s) should not be a member of %s",
				member, role,
			)
		}
		return nil
	}
}

func testAccCheckPostgresqlDatabaseDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "postgresql_database" {
			continue
		}

		exists, err := checkDatabaseExists(client, rs.Primary.ID)

		if err != nil {
			return fmt.Errorf("Error checking db %s", err)
		}

		if exists {
			return errors.New("Db still exists after destroy")
		}
	}

	return nil
}

func testAccCheckPostgresqlDatabaseExists(n string, searchPath []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Resource not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No ID is set")
		}

		client := testAccProvider.Meta().(*Client)
		exists, err := checkDatabaseExists(client, rs.Primary.ID)

		if err != nil {
			return fmt.Errorf("Error checking db %s", err)
		}

		if !exists {
			return errors.New("Db not found")
		}

		if searchPath != nil {
			if err := checkDBSearchPath(client, rs.Primary.ID, searchPath); err != nil {
				return fmt.Errorf("Error checking db search_path %s", err)
			}
		}

		return nil
	}
}

func checkDatabaseExists(client *Client, dbName string) (bool, error) {
	db, err := client.Connect()
	if err != nil {
		return false, err
	}
	var _rez int
	err = db.QueryRow("SELECT 1 from pg_database d WHERE datname=$1", dbName).Scan(&_rez)
	switch {
	case err == sql.ErrNoRows:
		return false, nil
	case err != nil:
		return false, fmt.Errorf("Error reading info about database: %s", err)
	}

	return true, nil
}

func checkDBSearchPath(client *Client, dbName string, expectedSearchPath []string) error {
	db, err := client.Connect()
	if err != nil {
		return err
	}

	var searchPathStr string
	err = db.QueryRow(`
		    SELECT (pg_options_to_table(drs.setconfig)).option_value 
		    FROM pg_catalog.pg_database AS d, pg_catalog.pg_db_role_setting AS drs 
		    WHERE d.datname = $1 AND d.oid = drs.setdatabase`,
		dbName,
	).Scan(&searchPathStr)

	// The query returns ErrNoRows if the search path hasn't been altered.
	if err != nil && err == sql.ErrNoRows {
		searchPathStr = ""
	} else if err != nil {
		return fmt.Errorf("Error reading search_path: %v", err)
	}

	searchPath := strings.Split(searchPathStr, ", ")
	for i := range searchPath {
		searchPath[i] = strings.Trim(searchPath[i], `"`)
	}
	sort.Strings(expectedSearchPath)
	if !reflect.DeepEqual(searchPath, expectedSearchPath) {
		return fmt.Errorf(
			"search_path is not equal to expected value. expected %v - got %v",
			expectedSearchPath, searchPath,
		)
	}
	return nil
}

var testAccPostgreSQLDatabaseConfig = `
resource "postgresql_role" "myrole" {
  name = "myrole"
  login = true
}

resource "postgresql_database" "mydb" {
   name = "mydb"
   owner = "${postgresql_role.myrole.name}"
}

resource "postgresql_database" "mydb2" {
   name = "mydb2"
   owner = "${postgresql_role.myrole.name}"
}

resource "postgresql_database" "default_opts" {
   name = "default_opts_name"
   owner = "${postgresql_role.myrole.name}"
   template = "template0"
   encoding = "UTF8"
   lc_collate = "C"
   lc_ctype = "C"
   connection_limit = -1
   is_template = false
}

resource "postgresql_database" "modified_opts" {
   name = "custom_template_db"
   owner = "${postgresql_role.myrole.name}"
   template = "template0"
   encoding = "UTF8"
   lc_collate = "en_US.UTF-8"
   lc_ctype = "en_US.UTF-8"
   connection_limit = 10
   is_template = true
}

resource "postgresql_database" "pathological_opts" {
   name = "bad_template_db"
   owner = "${postgresql_role.myrole.name}"
   template = "template0"
   encoding = "LATIN1"
   lc_collate = "C"
   lc_ctype = "C"
   connection_limit = 0
   is_template = true
}

resource "postgresql_database" "pg_default_opts" {
  lifecycle {
    ignore_changes = [
      "template",
      "encoding",
      "lc_collate",
      "lc_ctype",
      "tablespace_name",
    ]
  }

  name = "pg_defaults_db"
  owner = "${postgresql_role.myrole.name}"
  template = "DEFAULT"
  encoding = "DEFAULT"
  lc_collate = "DEFAULT"
  lc_ctype = "DEFAULT"
  tablespace_name = "DEFAULT"
  connection_limit = 0
  is_template = true
  search_path = []
}

resource "postgresql_database" "mydb_default_owner" {
   name = "mydb_default_owner"
}

resource "postgresql_database" "mydb_search_path" {
   name = "mydb_search_path"
   search_path = ["bar", "foo-with-hyphen"]
}

`
