package auth0

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccPrompt(t *testing.T) {
	httpRecorder := configureHTTPRecorder(t)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testProviders(httpRecorder),
		Steps: []resource.TestStep{
			{
				Config: testAccPromptCreate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_prompt.prompt", "universal_login_experience", "classic"),
					resource.TestCheckResourceAttr("auth0_prompt.prompt", "identifier_first", "false"),
				),
			},
			{
				Config: testAccPromptUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_prompt.prompt", "universal_login_experience", "new"),
					resource.TestCheckResourceAttr("auth0_prompt.prompt", "identifier_first", "true"),
				),
			},
		},
	})
}

const testAccPromptCreate = `
resource "auth0_prompt" "prompt" {
  universal_login_experience = "classic"
  identifier_first = false
}
`

const testAccPromptUpdate = `
resource "auth0_prompt" "prompt" {
  universal_login_experience = "new"
  identifier_first = true
}
`
