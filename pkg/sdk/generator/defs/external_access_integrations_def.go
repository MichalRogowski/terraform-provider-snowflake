package defs

import (
	g "github.com/Snowflake-Labs/terraform-provider-snowflake/pkg/sdk/generator/gen"

	"github.com/Snowflake-Labs/terraform-provider-snowflake/pkg/sdk/generator/gen/sdkcommons"
)

var externalAccessIntegrationAllowedAuthenticationSecretsDef = g.NewQueryStruct("ExternalAccessIntegrationAllowedAuthenticationSecrets").
	OptionalSQLWithCustomFieldName("AllSecrets", "ALL").
	OptionalSQLWithCustomFieldName("NoSecrets", "NONE").
	List("AllowedList", g.KindOfT[sdkcommons.SchemaObjectIdentifier](), g.ListOptions().Parentheses()).
	WithValidation(g.ExactlyOneValueSet, "AllSecrets", "NoSecrets", "AllowedList")

var externalAccessIntegrationAllowedApiAuthenticationIntegrationsDef = g.NewQueryStruct("ExternalAccessIntegrationAllowedApiAuthenticationIntegrations").
	OptionalSQLWithCustomFieldName("NoIntegrations", "NONE").
	List("AllowedList", g.KindOfT[sdkcommons.AccountObjectIdentifier](), g.ListOptions().Parentheses()).
	WithValidation(g.ExactlyOneValueSet, "NoIntegrations", "AllowedList")

var externalAccessIntegrationsDef = g.NewInterface(
	"ExternalAccessIntegrations",
	"ExternalAccessIntegration",
	g.KindOfT[sdkcommons.AccountObjectIdentifier](),
).
	CreateOperation(
		"https://docs.snowflake.com/en/sql-reference/sql/create-external-access-integration",
		g.NewQueryStruct("CreateExternalAccessIntegration").
			Create().
			OrReplace().
			SQL("EXTERNAL ACCESS INTEGRATION").
			Name(). // no IfNotExists() -- not supported by this object's grammar
			ListAssignment("ALLOWED_NETWORK_RULES", g.KindOfT[sdkcommons.SchemaObjectIdentifier](), g.ParameterOptions().Parentheses().Required()).
			OptionalQueryStructField(
				"AllowedApiAuthenticationIntegrations",
				externalAccessIntegrationAllowedApiAuthenticationIntegrationsDef,
				g.KeywordOptions().SQL("ALLOWED_API_AUTHENTICATION_INTEGRATIONS ="),
			).
			OptionalQueryStructField(
				"AllowedAuthenticationSecrets",
				externalAccessIntegrationAllowedAuthenticationSecretsDef,
				g.KeywordOptions().SQL("ALLOWED_AUTHENTICATION_SECRETS ="),
			).
			BooleanAssignment("ENABLED", g.ParameterOptions().Required()).
			OptionalComment().
			WithValidation(g.ValidIdentifier, "name"),
		externalAccessIntegrationAllowedApiAuthenticationIntegrationsDef,
		externalAccessIntegrationAllowedAuthenticationSecretsDef,
	).
	AlterOperation(
		"https://docs.snowflake.com/en/sql-reference/sql/alter-external-access-integration",
		g.NewQueryStruct("AlterExternalAccessIntegration").
			Alter().
			SQL("EXTERNAL ACCESS INTEGRATION").
			IfExists().
			Name().
			OptionalQueryStructField(
				"Set",
				g.NewQueryStruct("ExternalAccessIntegrationSet").
					ListAssignment("ALLOWED_NETWORK_RULES", g.KindOfT[sdkcommons.SchemaObjectIdentifier](), g.ParameterOptions().Parentheses()).
					OptionalQueryStructField(
						"AllowedApiAuthenticationIntegrations",
						externalAccessIntegrationAllowedApiAuthenticationIntegrationsDef,
						g.KeywordOptions().SQL("ALLOWED_API_AUTHENTICATION_INTEGRATIONS ="),
					).
					OptionalQueryStructField(
						"AllowedAuthenticationSecrets",
						externalAccessIntegrationAllowedAuthenticationSecretsDef,
						g.KeywordOptions().SQL("ALLOWED_AUTHENTICATION_SECRETS ="),
					).
					OptionalBooleanAssignment("ENABLED", g.ParameterOptions()).
					OptionalComment().
					WithValidation(g.AtLeastOneValueSet, "AllowedNetworkRules", "AllowedApiAuthenticationIntegrations", "AllowedAuthenticationSecrets", "Enabled", "Comment"),
				g.KeywordOptions().SQL("SET"),
			).
			OptionalQueryStructField(
				"Unset",
				g.NewQueryStruct("ExternalAccessIntegrationUnset").
					OptionalSQL("ALLOWED_NETWORK_RULES").
					OptionalSQL("ALLOWED_API_AUTHENTICATION_INTEGRATIONS").
					OptionalSQL("ALLOWED_AUTHENTICATION_SECRETS").
					OptionalSQL("COMMENT"). // no Enabled -- not in this object's UNSET grammar
					WithValidation(g.AtLeastOneValueSet, "AllowedNetworkRules", "AllowedApiAuthenticationIntegrations", "AllowedAuthenticationSecrets", "Comment"),
				g.ListOptions().NoParentheses().SQL("UNSET"),
			).
			OptionalSetTags().
			OptionalUnsetTags().
			WithValidation(g.ValidIdentifier, "name").
			WithValidation(g.ExactlyOneValueSet, "Set", "Unset", "SetTags", "UnsetTags"),
	).
	DropOperation(
		"https://docs.snowflake.com/en/sql-reference/sql/drop-integration",
		g.NewQueryStruct("DropExternalAccessIntegration").
			Drop().
			SQL("EXTERNAL ACCESS INTEGRATION").
			IfExists().
			Name().
			WithValidation(g.ValidIdentifier, "name"),
	).
	ShowOperationWithPairedStructs(
		"https://docs.snowflake.com/en/sql-reference/sql/show-integrations",
		g.StructPair("showExternalAccessIntegrationsDbRow", "ExternalAccessIntegration").
			Text("name").
			Text("type", g.WithPlainFieldName("ExternalAccessType")).
			Text("category").
			Bool("enabled").
			OptionalText("comment", g.WithRequiredInPlain()).
			Time("created_on"),
		g.NewQueryStruct("ShowExternalAccessIntegrations").
			Show().
			SQL("EXTERNAL ACCESS INTEGRATIONS").
			OptionalLike(),
	).
	DescribeOperationWithPairedStructs(
		g.DescriptionMappingKindSlice,
		"https://docs.snowflake.com/en/sql-reference/sql/desc-integration",
		g.StructPair("descExternalAccessIntegrationsDbRow", "ExternalAccessIntegrationProperty").
			Text("property", g.WithPlainFieldName("Name")).
			Text("property_type", g.WithPlainFieldName("Type")).
			Text("property_value", g.WithPlainFieldName("Value")).
			Text("property_default", g.WithPlainFieldName("Default")),
		g.NewQueryStruct("DescribeExternalAccessIntegration").
			Describe().
			SQL("EXTERNAL ACCESS INTEGRATION").
			Name().
			WithValidation(g.ValidIdentifier, "name"),
	).
	WithShowObjectType("Integration")
