//go:build non_account_level_tests

package testacc

import (
	"testing"

	r "github.com/Snowflake-Labs/terraform-provider-snowflake/pkg/resources"

	"github.com/Snowflake-Labs/terraform-provider-snowflake/pkg/acceptance/bettertestspoc/assert/resourceassert"
	"github.com/Snowflake-Labs/terraform-provider-snowflake/pkg/acceptance/bettertestspoc/assert/resourceshowoutputassert"
	"github.com/Snowflake-Labs/terraform-provider-snowflake/pkg/acceptance/bettertestspoc/config"
	"github.com/Snowflake-Labs/terraform-provider-snowflake/pkg/acceptance/bettertestspoc/config/model"
	"github.com/Snowflake-Labs/terraform-provider-snowflake/pkg/acceptance/helpers/random"
	"github.com/Snowflake-Labs/terraform-provider-snowflake/pkg/provider/resources"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

func TestAcc_ExternalAccessIntegration_basic(t *testing.T) {
	networkRule, networkRuleCleanup := testClient().NetworkRule.Create(t)
	t.Cleanup(networkRuleCleanup)
	networkRuleId := networkRule.ID().FullyQualifiedName()

	id := testClient().Ids.RandomAccountObjectIdentifier()
	comment := random.Comment()

	basicModel := model.ExternalAccessIntegration("test", id.Name(), []string{networkRuleId}, false)
	updatedModel := model.ExternalAccessIntegration("test", id.Name(), []string{networkRuleId}, true).
		WithComment(comment)

	ref := basicModel.ResourceReference()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.RequireAbove(tfversion.Version1_5_0),
		},
		CheckDestroy: CheckDestroy(t, resources.ExternalAccessIntegration),
		Steps: []resource.TestStep{
			// Create with only required attributes
			{
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(ref, plancheck.ResourceActionCreate),
					},
				},
				Config: config.FromModels(t, basicModel),
				Check: assertThat(
					t,
					resourceassert.ExternalAccessIntegrationResource(t, ref).
						HasNameString(id.Name()).
						HasEnabledString(r.BooleanFalse).
						HasAllowedNetworkRules(networkRuleId).
						HasAllowedApiAuthenticationIntegrationsEmpty().
						HasAllowedAuthenticationSecretsEmpty().
						HasCommentEmpty().
						HasFullyQualifiedNameString(id.FullyQualifiedName()),
					resourceshowoutputassert.ExternalAccessIntegrationShowOutput(t, ref).
						HasName(id.Name()).
						HasCategory("EXTERNAL_ACCESS").
						HasEnabled(false).
						HasComment(""),
				),
			},
			// Import
			{
				Config:            config.FromModels(t, basicModel),
				ResourceName:      ref,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update: enable and set comment
			{
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(ref, plancheck.ResourceActionUpdate),
					},
				},
				Config: config.FromModels(t, updatedModel),
				Check: assertThat(
					t,
					resourceassert.ExternalAccessIntegrationResource(t, ref).
						HasNameString(id.Name()).
						HasEnabledString(r.BooleanTrue).
						HasAllowedNetworkRules(networkRuleId).
						HasCommentString(comment),
					resourceshowoutputassert.ExternalAccessIntegrationShowOutput(t, ref).
						HasEnabled(true).
						HasComment(comment),
				),
			},
			// Unset comment (back to basic)
			{
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(ref, plancheck.ResourceActionUpdate),
					},
				},
				Config: config.FromModels(t, basicModel),
				Check: assertThat(
					t,
					resourceassert.ExternalAccessIntegrationResource(t, ref).
						HasEnabledString(r.BooleanFalse).
						HasCommentEmpty(),
				),
			},
		},
	})
}

func TestAcc_ExternalAccessIntegration_complete(t *testing.T) {
	networkRule, networkRuleCleanup := testClient().NetworkRule.Create(t)
	t.Cleanup(networkRuleCleanup)
	networkRuleId := networkRule.ID().FullyQualifiedName()

	secretId, secretCleanup := testClient().Secret.CreateRandomPasswordSecret(t)
	t.Cleanup(secretCleanup)

	apiAuthIntegration, apiAuthIntegrationCleanup := testClient().SecurityIntegration.CreateApiAuthenticationWithClientCredentialsFlow(t)
	t.Cleanup(apiAuthIntegrationCleanup)
	apiAuthIntegrationId := apiAuthIntegration.ID().FullyQualifiedName()

	id := testClient().Ids.RandomAccountObjectIdentifier()
	comment := random.Comment()

	completeModel := model.ExternalAccessIntegration("test", id.Name(), []string{networkRuleId}, true).
		WithAllowedApiAuthenticationIntegrations([]string{apiAuthIntegrationId}).
		WithAllowedAuthenticationSecrets([]string{secretId.FullyQualifiedName()}).
		WithComment(comment)

	basicModel := model.ExternalAccessIntegration("test", id.Name(), []string{networkRuleId}, true)

	ref := completeModel.ResourceReference()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.RequireAbove(tfversion.Version1_5_0),
		},
		CheckDestroy: CheckDestroy(t, resources.ExternalAccessIntegration),
		Steps: []resource.TestStep{
			// Create with all attributes
			{
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(ref, plancheck.ResourceActionCreate),
					},
				},
				Config: config.FromModels(t, completeModel),
				Check: assertThat(
					t,
					resourceassert.ExternalAccessIntegrationResource(t, ref).
						HasNameString(id.Name()).
						HasEnabledString(r.BooleanTrue).
						HasAllowedNetworkRules(networkRuleId).
						HasAllowedApiAuthenticationIntegrations(apiAuthIntegrationId).
						HasAllowedAuthenticationSecrets(secretId.FullyQualifiedName()).
						HasCommentString(comment).
						HasFullyQualifiedNameString(id.FullyQualifiedName()),
					resourceshowoutputassert.ExternalAccessIntegrationShowOutput(t, ref).
						HasName(id.Name()).
						HasCategory("EXTERNAL_ACCESS").
						HasEnabled(true).
						HasComment(comment),
				),
			},
			// Import
			{
				Config:            config.FromModels(t, completeModel),
				ResourceName:      ref,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Unset the optional lists and comment
			{
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(ref, plancheck.ResourceActionUpdate),
					},
				},
				Config: config.FromModels(t, basicModel),
				Check: assertThat(
					t,
					resourceassert.ExternalAccessIntegrationResource(t, ref).
						HasAllowedNetworkRules(networkRuleId).
						HasAllowedApiAuthenticationIntegrationsEmpty().
						HasAllowedAuthenticationSecretsEmpty().
						HasCommentEmpty(),
				),
			},
		},
	})
}
