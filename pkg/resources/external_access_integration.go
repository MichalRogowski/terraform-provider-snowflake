package resources

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"github.com/Snowflake-Labs/terraform-provider-snowflake/pkg/helpers"
	"github.com/Snowflake-Labs/terraform-provider-snowflake/pkg/internal/collections"
	"github.com/Snowflake-Labs/terraform-provider-snowflake/pkg/internal/provider"
	"github.com/Snowflake-Labs/terraform-provider-snowflake/pkg/provider/previewfeatures"
	"github.com/Snowflake-Labs/terraform-provider-snowflake/pkg/provider/resources"
	"github.com/Snowflake-Labs/terraform-provider-snowflake/pkg/schemas"
	"github.com/Snowflake-Labs/terraform-provider-snowflake/pkg/sdk"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var externalAccessIntegrationSchema = map[string]*schema.Schema{
	"name": {
		Type:             schema.TypeString,
		Required:         true,
		ForceNew:         true,
		Description:      blocklistedCharactersFieldDescription("Specifies the identifier for the external access integration; must be unique in your account. Due to technical limitations (read more [here](../guides/identifiers_rework_design_decisions#known-limitations-and-identifier-recommendations)), avoid using the following characters: `|`, `.`, `\"`."),
		DiffSuppressFunc: suppressIdentifierQuoting,
	},
	"allowed_network_rules": {
		Type: schema.TypeSet,
		Elem: &schema.Schema{
			Type:             schema.TypeString,
			ValidateDiagFunc: IsValidIdentifier[sdk.SchemaObjectIdentifier](),
		},
		DiffSuppressFunc: NormalizeAndCompareIdentifiersInSet("allowed_network_rules"),
		Required:         true,
		Description:      relatedResourceDescription("Specifies the network rules that contain the network identifiers that are allowed to be used by the external access integration.", resources.NetworkRule),
	},
	"allowed_api_authentication_integrations": {
		Type: schema.TypeSet,
		Elem: &schema.Schema{
			Type:             schema.TypeString,
			ValidateDiagFunc: IsValidIdentifier[sdk.AccountObjectIdentifier](),
		},
		DiffSuppressFunc: NormalizeAndCompareIdentifiersInSet("allowed_api_authentication_integrations"),
		Optional:         true,
		Description:      "Specifies the security integrations for external API authentication that are allowed to be used by the external access integration.",
	},
	"allowed_authentication_secrets": {
		Type: schema.TypeSet,
		Elem: &schema.Schema{
			Type:             schema.TypeString,
			ValidateDiagFunc: IsValidIdentifier[sdk.SchemaObjectIdentifier](),
		},
		DiffSuppressFunc: NormalizeAndCompareIdentifiersInSet("allowed_authentication_secrets"),
		Optional:         true,
		Description:      "Specifies the secrets that are allowed to be used to authenticate to the external network location referenced by the external access integration.",
	},
	"enabled": {
		Type:        schema.TypeBool,
		Required:    true,
		Description: "Specifies whether this external access integration is enabled or disabled. Only an enabled integration can be used in functions and procedures.",
	},
	"comment": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "Specifies a comment for the external access integration.",
	},
	FullyQualifiedNameAttributeName: schemas.FullyQualifiedNameSchema,
	ShowOutputAttributeName: {
		Type:        schema.TypeList,
		Computed:    true,
		Description: "Outputs the result of `SHOW EXTERNAL ACCESS INTEGRATIONS` for the given external access integration.",
		Elem: &schema.Resource{
			Schema: schemas.ShowExternalAccessIntegrationSchema,
		},
	},
	DescribeOutputAttributeName: {
		Type:        schema.TypeList,
		Computed:    true,
		Description: "Outputs the result of `DESCRIBE EXTERNAL ACCESS INTEGRATION` for the given external access integration.",
		Elem: &schema.Resource{
			Schema: schemas.DescribeExternalAccessIntegrationSchema,
		},
	},
}

func ExternalAccessIntegration() *schema.Resource {
	deleteFunc := ResourceDeleteContextFunc(
		sdk.ParseAccountObjectIdentifier,
		func(client *sdk.Client) DropSafelyFunc[sdk.AccountObjectIdentifier] {
			return client.ExternalAccessIntegrations.DropSafely
		},
	)

	return &schema.Resource{
		Schema: externalAccessIntegrationSchema,

		CreateContext: PreviewFeatureCreateContextWrapper(string(previewfeatures.ExternalAccessIntegrationResource), TrackingCreateWrapper(resources.ExternalAccessIntegration, CreateContextExternalAccessIntegration)),
		ReadContext:   PreviewFeatureReadContextWrapper(string(previewfeatures.ExternalAccessIntegrationResource), TrackingReadWrapper(resources.ExternalAccessIntegration, ReadContextExternalAccessIntegration)),
		UpdateContext: PreviewFeatureUpdateContextWrapper(string(previewfeatures.ExternalAccessIntegrationResource), TrackingUpdateWrapper(resources.ExternalAccessIntegration, UpdateContextExternalAccessIntegration)),
		DeleteContext: PreviewFeatureDeleteContextWrapper(string(previewfeatures.ExternalAccessIntegrationResource), TrackingDeleteWrapper(resources.ExternalAccessIntegration, deleteFunc)),
		Description:   "Resource used to manage external access integrations. For more information, check [external access integrations documentation](https://docs.snowflake.com/en/sql-reference/sql/create-external-access-integration).",

		CustomizeDiff: TrackingCustomDiffWrapper(resources.ExternalAccessIntegration, customdiff.All(
			ComputedIfAnyAttributeChanged(externalAccessIntegrationSchema, ShowOutputAttributeName, "name", "enabled", "comment"),
			ComputedIfAnyAttributeChanged(externalAccessIntegrationSchema, DescribeOutputAttributeName, "name", "enabled", "comment", "allowed_network_rules", "allowed_api_authentication_integrations", "allowed_authentication_secrets"),
			ComputedIfAnyAttributeChanged(externalAccessIntegrationSchema, FullyQualifiedNameAttributeName, "name"),
		)),

		Importer: &schema.ResourceImporter{
			StateContext: TrackingImportWrapper(resources.ExternalAccessIntegration, ImportName[sdk.AccountObjectIdentifier]),
		},
		Timeouts: defaultTimeouts,
	}
}

func CreateContextExternalAccessIntegration(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*provider.Context).Client

	id, err := sdk.ParseAccountObjectIdentifier(d.Get("name").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	allowedNetworkRules, err := parseSchemaObjectIdentifierSet(d.Get("allowed_network_rules"))
	if err != nil {
		return diag.FromErr(err)
	}

	req := sdk.NewCreateExternalAccessIntegrationRequest(id, allowedNetworkRules, d.Get("enabled").(bool))

	if v, ok := d.GetOk("allowed_api_authentication_integrations"); ok {
		allowedApiAuthenticationIntegrations, err := parseAccountObjectIdentifierSet(v)
		if err != nil {
			return diag.FromErr(err)
		}
		req.WithAllowedApiAuthenticationIntegrations(*sdk.NewExternalAccessIntegrationAllowedApiAuthenticationIntegrationsRequest().WithAllowedList(allowedApiAuthenticationIntegrations))
	}

	if v, ok := d.GetOk("allowed_authentication_secrets"); ok {
		allowedAuthenticationSecrets, err := parseSchemaObjectIdentifierSet(v)
		if err != nil {
			return diag.FromErr(err)
		}
		req.WithAllowedAuthenticationSecrets(*sdk.NewExternalAccessIntegrationAllowedAuthenticationSecretsRequest().WithAllowedList(allowedAuthenticationSecrets))
	}

	if v, ok := d.GetOk("comment"); ok {
		req.WithComment(v.(string))
	}

	if err := client.ExternalAccessIntegrations.Create(ctx, req); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(helpers.EncodeResourceIdentifier(id))

	return ReadContextExternalAccessIntegration(ctx, d, meta)
}

func ReadContextExternalAccessIntegration(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*provider.Context).Client

	id, err := sdk.ParseAccountObjectIdentifier(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	integration, err := client.ExternalAccessIntegrations.ShowByIDSafely(ctx, id)
	if err != nil {
		if errors.Is(err, sdk.ErrObjectNotFound) {
			d.SetId("")
			return diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.Warning,
					Summary:  "Failed to query external access integration. Marking the resource as removed.",
					Detail:   fmt.Sprintf("External access integration id: %s, Err: %s", id.FullyQualifiedName(), err),
				},
			}
		}
		return diag.FromErr(err)
	}

	properties, err := client.ExternalAccessIntegrations.Describe(ctx, id)
	if err != nil {
		return diag.FromErr(err)
	}

	allowedNetworkRules, err := externalAccessIntegrationIdentifierListFromProperties(properties, "ALLOWED_NETWORK_RULES")
	if err != nil {
		return diag.FromErr(err)
	}
	allowedApiAuthenticationIntegrations, err := externalAccessIntegrationIdentifierListFromProperties(properties, "ALLOWED_API_AUTHENTICATION_INTEGRATIONS")
	if err != nil {
		return diag.FromErr(err)
	}
	allowedAuthenticationSecrets, err := externalAccessIntegrationIdentifierListFromProperties(properties, "ALLOWED_AUTHENTICATION_SECRETS")
	if err != nil {
		return diag.FromErr(err)
	}

	if errs := errors.Join(
		d.Set("name", id.Name()),
		d.Set("enabled", integration.Enabled),
		d.Set("comment", integration.Comment),
		d.Set("allowed_network_rules", allowedNetworkRules),
		d.Set("allowed_api_authentication_integrations", allowedApiAuthenticationIntegrations),
		d.Set("allowed_authentication_secrets", allowedAuthenticationSecrets),
		d.Set(FullyQualifiedNameAttributeName, id.FullyQualifiedName()),
		d.Set(ShowOutputAttributeName, []map[string]any{schemas.ExternalAccessIntegrationToSchema(integration)}),
		d.Set(DescribeOutputAttributeName, []map[string]any{schemas.ExternalAccessIntegrationPropertiesToSchema(properties)}),
	); errs != nil {
		return diag.FromErr(errs)
	}

	return nil
}

func UpdateContextExternalAccessIntegration(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*provider.Context).Client

	id, err := sdk.ParseAccountObjectIdentifier(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	set, unset := sdk.NewExternalAccessIntegrationSetRequest(), sdk.NewExternalAccessIntegrationUnsetRequest()

	// allowed_network_rules is required, so it can only be set (never unset).
	if d.HasChange("allowed_network_rules") {
		allowedNetworkRules, err := parseSchemaObjectIdentifierSet(d.Get("allowed_network_rules"))
		if err != nil {
			return diag.FromErr(err)
		}
		set.WithAllowedNetworkRules(allowedNetworkRules)
	}

	if d.HasChange("allowed_api_authentication_integrations") {
		if v, ok := d.GetOk("allowed_api_authentication_integrations"); ok {
			allowedApiAuthenticationIntegrations, err := parseAccountObjectIdentifierSet(v)
			if err != nil {
				return diag.FromErr(err)
			}
			set.WithAllowedApiAuthenticationIntegrations(*sdk.NewExternalAccessIntegrationAllowedApiAuthenticationIntegrationsRequest().WithAllowedList(allowedApiAuthenticationIntegrations))
		} else {
			unset.WithAllowedApiAuthenticationIntegrations(true)
		}
	}

	if d.HasChange("allowed_authentication_secrets") {
		if v, ok := d.GetOk("allowed_authentication_secrets"); ok {
			allowedAuthenticationSecrets, err := parseSchemaObjectIdentifierSet(v)
			if err != nil {
				return diag.FromErr(err)
			}
			set.WithAllowedAuthenticationSecrets(*sdk.NewExternalAccessIntegrationAllowedAuthenticationSecretsRequest().WithAllowedList(allowedAuthenticationSecrets))
		} else {
			unset.WithAllowedAuthenticationSecrets(true)
		}
	}

	if d.HasChange("enabled") {
		set.WithEnabled(d.Get("enabled").(bool))
	}

	if d.HasChange("comment") {
		if v, ok := d.GetOk("comment"); ok {
			set.WithComment(v.(string))
		} else {
			unset.WithComment(true)
		}
	}

	if !reflect.DeepEqual(set, sdk.NewExternalAccessIntegrationSetRequest()) {
		if err := client.ExternalAccessIntegrations.Alter(ctx, sdk.NewAlterExternalAccessIntegrationRequest(id).WithSet(*set)); err != nil {
			return diag.FromErr(err)
		}
	}

	if !reflect.DeepEqual(unset, sdk.NewExternalAccessIntegrationUnsetRequest()) {
		if err := client.ExternalAccessIntegrations.Alter(ctx, sdk.NewAlterExternalAccessIntegrationRequest(id).WithUnset(*unset)); err != nil {
			return diag.FromErr(err)
		}
	}

	return ReadContextExternalAccessIntegration(ctx, d, meta)
}

// externalAccessIntegrationIdentifierListFromProperties reads a DESCRIBE property that holds a list of object
// references and returns their fully qualified names. Snowflake returns these lists as a JSON array of objects,
// each holding a single fully qualified name field (the same representation used for network policies' network
// rules). We tolerate an empty/missing property by returning an empty slice.
func externalAccessIntegrationIdentifierListFromProperties(properties []sdk.ExternalAccessIntegrationProperty, propertyName string) ([]string, error) {
	property, err := collections.FindFirst(properties, func(p sdk.ExternalAccessIntegrationProperty) bool { return p.Name == propertyName })
	if err != nil {
		return []string{}, nil
	}
	return parseExternalAccessIntegrationObjectReferenceList(property.Value)
}

// parseExternalAccessIntegrationObjectReferenceList parses the Snowflake DESCRIBE representation of an object
// reference list into fully qualified names. The value is a JSON array of objects, each with a single string
// field holding the fully qualified name, e.g. `[{"fullyQualifiedRuleName":"DB.SCHEMA.RULE"}]`.
func parseExternalAccessIntegrationObjectReferenceList(value string) ([]string, error) {
	result := make([]string, 0)
	if value == "" {
		return result, nil
	}
	var entries []map[string]string
	if err := json.Unmarshal([]byte(value), &entries); err != nil {
		return nil, fmt.Errorf("failed to parse object reference list %q: %w", value, err)
	}
	for _, entry := range entries {
		for _, name := range entry {
			result = append(result, name)
		}
	}
	return result, nil
}
