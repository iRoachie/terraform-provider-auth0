package auth0

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"testing"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/auth0/terraform-provider-auth0/auth0/internal/template"
)

func init() {
	resource.AddTestSweepers("auth0_user", &resource.Sweeper{
		Name: "auth0_user",
		F: func(_ string) error {
			api, err := Auth0()
			if err != nil {
				return err
			}

			var page int
			var result *multierror.Error
			for {
				userList, err := api.User.Search(
					management.Page(page),
					management.Query(`email.domain:"acceptance.test.com"`))
				if err != nil {
					return err
				}

				for _, user := range userList.Users {
					result = multierror.Append(
						result,
						api.User.Delete(user.GetID()),
					)
					log.Printf("[DEBUG] ✗ %s", user.GetName())
				}
				if !userList.HasNext() {
					break
				}
				page++
			}

			return result.ErrorOrNil()
		},
	})
}

func TestAccUserMissingRequiredParams(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: testProviders(nil),
		Steps: []resource.TestStep{
			{
				Config:      "resource auth0_user user {}",
				ExpectError: regexp.MustCompile(`The argument "connection_name" is required`),
			},
		},
	})
}

func TestAccUser(t *testing.T) {
	httpRecorder := configureHTTPRecorder(t)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testProviders(httpRecorder),
		Steps: []resource.TestStep{
			{
				Config: template.ParseTestName(testAccUserCreate, strings.ToLower(t.Name())),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_user.user", "user_id", fmt.Sprintf("auth0|%s", strings.ToLower(t.Name()))),
					resource.TestCheckResourceAttr("auth0_user.user", "email", fmt.Sprintf("%s@acceptance.test.com", strings.ToLower(t.Name()))),
					resource.TestCheckResourceAttr("auth0_user.user", "name", "Firstname Lastname"),
					resource.TestCheckResourceAttr("auth0_user.user", "family_name", "Lastname"),
					resource.TestCheckResourceAttr("auth0_user.user", "given_name", "Firstname"),
					resource.TestCheckResourceAttr("auth0_user.user", "nickname", strings.ToLower(t.Name())),
					resource.TestCheckResourceAttr("auth0_user.user", "connection_name", "Username-Password-Authentication"),
					resource.TestCheckResourceAttr("auth0_user.user", "roles.#", "0"),
					resource.TestCheckResourceAttr("auth0_user.user", "picture", "https://www.example.com/picture.jpg"),
				),
			},
			{
				Config: template.ParseTestName(testAccUserAddRole, strings.ToLower(t.Name())),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_user.user", "roles.#", "2"),
					resource.TestCheckResourceAttr("auth0_role.owner", "name", "owner"),
					resource.TestCheckResourceAttr("auth0_role.admin", "name", "admin"),
				),
			},
			{
				Config: template.ParseTestName(testAccUserRemoveRole, strings.ToLower(t.Name())),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_user.user", "roles.#", "1"),
				),
			},
		},
	})
}

const testAccUserCreate = `
resource auth0_user user {
	connection_name = "Username-Password-Authentication"
	username = "{{.testName}}"
	user_id = "{{.testName}}"
	email = "{{.testName}}@acceptance.test.com"
	password = "passpass$12$12"
	name = "Firstname Lastname"
	given_name = "Firstname"
	family_name = "Lastname"
	nickname = "{{.testName}}"
	picture = "https://www.example.com/picture.jpg"
	user_metadata = <<EOF
{
  "foo": "bar",
  "bar": { "baz": "qux" }
}
EOF
	app_metadata = <<EOF
{
  "foo": "bar",
  "bar": { "baz": "qux" }
}
EOF
}
`

const testAccUserAddRole = `
resource auth0_user user {
	depends_on = [auth0_role.owner, auth0_role.admin]
	connection_name = "Username-Password-Authentication"
	username = "{{.testName}}"
	user_id = "{{.testName}}"
	email = "{{.testName}}@acceptance.test.com"
	password = "passpass$12$12"
	name = "Firstname Lastname"
	given_name = "Firstname"
	family_name = "Lastname"
	nickname = "{{.testName}}"
	picture = "https://www.example.com/picture.jpg"
	roles = [ auth0_role.owner.id, auth0_role.admin.id ]
	user_metadata = <<EOF
{
  "foo": "bar",
  "bar": { "baz": "qux" }
}
EOF
app_metadata = <<EOF
{
  "foo": "bar",
  "bar": { "baz": "qux" }
}
EOF
}

resource auth0_role owner {
	name = "owner"
	description = "Owner"
}

resource auth0_role admin {
	name = "admin"
	description = "Administrator"
	depends_on = [auth0_role.owner]
}
`

const testAccUserRemoveRole = `
resource auth0_user user {
	depends_on = [auth0_role.admin]
	connection_name = "Username-Password-Authentication"
	username = "{{.testName}}"
	user_id = "{{.testName}}"
	email = "{{.testName}}@acceptance.test.com"
	password = "passpass$12$12"
	name = "Firstname Lastname"
	given_name = "Firstname"
	family_name = "Lastname"
	nickname = "{{.testName}}"
	picture = "https://www.example.com/picture.jpg"
	roles = [ auth0_role.admin.id ]
	user_metadata = <<EOF
{
  	"foo": "bar",
  	"bar": { "baz": "qux" }
}
EOF
  app_metadata = <<EOF
{
  	"foo": "bar",
  	"bar": { "baz": "qux" }
}
EOF
}

resource auth0_role admin {
	name = "admin"
	description = "Administrator"
}
`

func TestAccUserIssue218(t *testing.T) {
	httpRecorder := configureHTTPRecorder(t)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testProviders(httpRecorder),
		Steps: []resource.TestStep{
			{
				Config: template.ParseTestName(testAccUserIssue218, "issue#218"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_user.auth0_user_issue_218", "user_id", "auth0|id_issue#218"),
					resource.TestCheckResourceAttr("auth0_user.auth0_user_issue_218", "username", "user_issue#218"),
					resource.TestCheckResourceAttr("auth0_user.auth0_user_issue_218", "email", "issue.218.issue#218@acceptance.test.com"),
				),
			},
			{
				Config: template.ParseTestName(testAccUserIssue218, "issue#218"),
			},
		},
	})
}

const testAccUserIssue218 = `
resource auth0_user auth0_user_issue_218 {
  connection_name = "Username-Password-Authentication"
  user_id = "id_{{.testName}}"
  username = "user_{{.testName}}"
  email = "issue.218.{{.testName}}@acceptance.test.com"
  email_verified = true
  password = "MyPass123$"
}
`

func TestAccUserChangeUsername(t *testing.T) {
	httpRecorder := configureHTTPRecorder(t)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testProviders(httpRecorder),
		Steps: []resource.TestStep{
			{
				Config: template.ParseTestName(testAccUserChangeUsernameCreate, "terra"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_user.auth0_user_change_username", "username", "user_terra"),
					resource.TestCheckResourceAttr("auth0_user.auth0_user_change_username", "email", "change.username.terra@acceptance.test.com"),
					resource.TestCheckResourceAttr("auth0_user.auth0_user_change_username", "password", "MyPass123$"),
				),
			},
			{
				Config: template.ParseTestName(testAccUserChangeUsernameUpdate, "terra"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_user.auth0_user_change_username", "username", "user_x_terra"),
					resource.TestCheckResourceAttr("auth0_user.auth0_user_change_username", "email", "change.username.terra@acceptance.test.com"),
					resource.TestCheckResourceAttr("auth0_user.auth0_user_change_username", "password", "MyPass123$"),
				),
			},
			{
				Config:      template.ParseTestName(testAccUserChangeUsernameAndPassword, "terra"),
				ExpectError: regexp.MustCompile("cannot update username and password simultaneously"),
			},
		},
	})
}

const testAccUserChangeUsernameCreate = `
resource auth0_user auth0_user_change_username {
  connection_name = "Username-Password-Authentication"
  username = "user_{{.testName}}"
  email = "change.username.{{.testName}}@acceptance.test.com"
  email_verified = true
  password = "MyPass123$"
}
`

const testAccUserChangeUsernameUpdate = `
resource auth0_user auth0_user_change_username {
  connection_name = "Username-Password-Authentication"
  username = "user_x_{{.testName}}"
  email = "change.username.{{.testName}}@acceptance.test.com"
  email_verified = true
  password = "MyPass123$"
}
`

const testAccUserChangeUsernameAndPassword = `
resource auth0_user auth0_user_change_username {
  connection_name = "Username-Password-Authentication"
  username = "user_{{.testName}}"
  email = "change.username.{{.testName}}@acceptance.test.com"
  email_verified = true
  password = "MyPass123456$"
}
`
