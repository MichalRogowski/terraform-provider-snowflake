//go:build non_account_level_tests

package testint

import (
	"testing"

	"github.com/Snowflake-Labs/terraform-provider-snowflake/pkg/acceptance/bettertestspoc/assert/objectassert"
	"github.com/Snowflake-Labs/terraform-provider-snowflake/pkg/acceptance/helpers/random"
	"github.com/Snowflake-Labs/terraform-provider-snowflake/pkg/internal/collections"
	"github.com/Snowflake-Labs/terraform-provider-snowflake/pkg/sdk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInt_ExternalAccessIntegrations(t *testing.T) {
	client := testClient(t)
	ctx := testContext(t)

	// findProperty returns the described property row with the given name.
	findProperty := func(t *testing.T, props []sdk.ExternalAccessIntegrationProperty, name string) *sdk.ExternalAccessIntegrationProperty {
		t.Helper()
		prop, err := collections.FindFirst(props, func(p sdk.ExternalAccessIntegrationProperty) bool {
			return p.Name == name
		})
		require.NoError(t, err, "expected to find property %q", name)
		return prop
	}

	createNetworkRule := func(t *testing.T) sdk.SchemaObjectIdentifier {
		t.Helper()
		networkRule, networkRuleCleanup := testClientHelper().NetworkRule.Create(t)
		t.Cleanup(networkRuleCleanup)
		return networkRule.ID()
	}

	t.Run("create - basic", func(t *testing.T) {
		networkRuleId := createNetworkRule(t)
		id := testClientHelper().Ids.RandomAccountObjectIdentifier()

		err := client.ExternalAccessIntegrations.Create(ctx, sdk.NewCreateExternalAccessIntegrationRequest(id, []sdk.SchemaObjectIdentifier{networkRuleId}, true))
		require.NoError(t, err)
		t.Cleanup(testClientHelper().ExternalAccessIntegration.DropExternalAccessIntegrationFunc(t, id))

		integration, err := client.ExternalAccessIntegrations.ShowByID(ctx, id)
		require.NoError(t, err)

		assertThatObject(
			t, objectassert.ExternalAccessIntegrationFromObject(t, integration).
				HasName(id.Name()).
				HasExternalAccessType("EXTERNAL_ACCESS").
				HasCategory("EXTERNAL_ACCESS").
				HasEnabled(true).
				HasComment(""),
		)
	})

	t.Run("create - complete", func(t *testing.T) {
		networkRuleId := createNetworkRule(t)
		secretId, secretCleanup := testClientHelper().Secret.CreateRandomPasswordSecret(t)
		t.Cleanup(secretCleanup)
		id := testClientHelper().Ids.RandomAccountObjectIdentifier()
		comment := random.Comment()

		req := sdk.NewCreateExternalAccessIntegrationRequest(id, []sdk.SchemaObjectIdentifier{networkRuleId}, true).
			WithComment(comment).
			WithAllowedAuthenticationSecrets(*sdk.NewExternalAccessIntegrationAllowedAuthenticationSecretsRequest().
				WithAllowedList([]sdk.SchemaObjectIdentifier{secretId})).
			WithAllowedApiAuthenticationIntegrations(*sdk.NewExternalAccessIntegrationAllowedApiAuthenticationIntegrationsRequest().
				WithNoIntegrations(true))
		err := client.ExternalAccessIntegrations.Create(ctx, req)
		require.NoError(t, err)
		t.Cleanup(testClientHelper().ExternalAccessIntegration.DropExternalAccessIntegrationFunc(t, id))

		assertThatObject(
			t, objectassert.ExternalAccessIntegration(t, id).
				HasName(id.Name()).
				HasEnabled(true).
				HasComment(comment),
		)

		props, err := client.ExternalAccessIntegrations.Describe(ctx, id)
		require.NoError(t, err)
		require.NotEmpty(t, props)

		enabled := findProperty(t, props, "ENABLED")
		assert.Equal(t, "true", enabled.Value)
		networkRules := findProperty(t, props, "ALLOWED_NETWORK_RULES")
		assert.Contains(t, networkRules.Value, networkRuleId.Name())
		secrets := findProperty(t, props, "ALLOWED_AUTHENTICATION_SECRETS")
		assert.Contains(t, secrets.Value, secretId.Name())
		commentProp := findProperty(t, props, "COMMENT")
		assert.Equal(t, comment, commentProp.Value)
	})

	t.Run("alter - set", func(t *testing.T) {
		networkRuleId := createNetworkRule(t)
		id, cleanup := testClientHelper().ExternalAccessIntegration.CreateExternalAccessIntegration(t, networkRuleId)
		t.Cleanup(cleanup)
		comment := random.Comment()

		err := client.ExternalAccessIntegrations.Alter(ctx, sdk.NewAlterExternalAccessIntegrationRequest(id).
			WithSet(*sdk.NewExternalAccessIntegrationSetRequest().
				WithEnabled(false).
				WithComment(comment)))
		require.NoError(t, err)

		assertThatObject(
			t, objectassert.ExternalAccessIntegration(t, id).
				HasEnabled(false).
				HasComment(comment),
		)
	})

	t.Run("alter - unset", func(t *testing.T) {
		networkRuleId := createNetworkRule(t)
		id := testClientHelper().Ids.RandomAccountObjectIdentifier()
		comment := random.Comment()

		err := client.ExternalAccessIntegrations.Create(ctx, sdk.NewCreateExternalAccessIntegrationRequest(id, []sdk.SchemaObjectIdentifier{networkRuleId}, true).WithComment(comment))
		require.NoError(t, err)
		t.Cleanup(testClientHelper().ExternalAccessIntegration.DropExternalAccessIntegrationFunc(t, id))

		err = client.ExternalAccessIntegrations.Alter(ctx, sdk.NewAlterExternalAccessIntegrationRequest(id).
			WithUnset(*sdk.NewExternalAccessIntegrationUnsetRequest().WithComment(true)))
		require.NoError(t, err)

		assertThatObject(
			t, objectassert.ExternalAccessIntegration(t, id).
				HasComment(""),
		)
	})

	t.Run("drop", func(t *testing.T) {
		networkRuleId := createNetworkRule(t)
		id := testClientHelper().Ids.RandomAccountObjectIdentifier()

		err := client.ExternalAccessIntegrations.Create(ctx, sdk.NewCreateExternalAccessIntegrationRequest(id, []sdk.SchemaObjectIdentifier{networkRuleId}, true))
		require.NoError(t, err)

		err = client.ExternalAccessIntegrations.Drop(ctx, sdk.NewDropExternalAccessIntegrationRequest(id))
		require.NoError(t, err)

		_, err = client.ExternalAccessIntegrations.ShowByID(ctx, id)
		require.ErrorIs(t, err, collections.ErrObjectNotFound)
	})

	t.Run("drop with if exists", func(t *testing.T) {
		id := NonExistingAccountObjectIdentifier

		err := client.ExternalAccessIntegrations.Drop(ctx, sdk.NewDropExternalAccessIntegrationRequest(id))
		require.ErrorIs(t, err, sdk.ErrObjectNotExistOrAuthorized)

		err = client.ExternalAccessIntegrations.Drop(ctx, sdk.NewDropExternalAccessIntegrationRequest(id).WithIfExists(true))
		require.NoError(t, err)
	})

	t.Run("show with like", func(t *testing.T) {
		networkRuleId := createNetworkRule(t)
		id, cleanup := testClientHelper().ExternalAccessIntegration.CreateExternalAccessIntegration(t, networkRuleId)
		t.Cleanup(cleanup)

		integrations, err := client.ExternalAccessIntegrations.Show(ctx, sdk.NewShowExternalAccessIntegrationRequest().WithLike(sdk.Like{Pattern: sdk.String(id.Name())}))
		require.NoError(t, err)
		require.Len(t, integrations, 1)
		assert.Equal(t, id.Name(), integrations[0].Name)
	})

	t.Run("describe: non-existing", func(t *testing.T) {
		id := NonExistingAccountObjectIdentifier

		_, err := client.ExternalAccessIntegrations.Describe(ctx, id)
		assert.ErrorIs(t, err, sdk.ErrObjectNotExistOrAuthorized)
	})
}
