package model

import (
	tfconfig "github.com/hashicorp/terraform-plugin-testing/config"

	"github.com/Snowflake-Labs/terraform-provider-snowflake/pkg/internal/collections"
)

func (e *ExternalAccessIntegrationModel) WithAllowedNetworkRules(allowedNetworkRules []string) *ExternalAccessIntegrationModel {
	return e.WithAllowedNetworkRulesValue(stringSetVariable(allowedNetworkRules))
}

func (e *ExternalAccessIntegrationModel) WithAllowedApiAuthenticationIntegrations(allowedApiAuthenticationIntegrations []string) *ExternalAccessIntegrationModel {
	return e.WithAllowedApiAuthenticationIntegrationsValue(stringSetVariable(allowedApiAuthenticationIntegrations))
}

func (e *ExternalAccessIntegrationModel) WithAllowedAuthenticationSecrets(allowedAuthenticationSecrets []string) *ExternalAccessIntegrationModel {
	return e.WithAllowedAuthenticationSecretsValue(stringSetVariable(allowedAuthenticationSecrets))
}

func stringSetVariable(values []string) tfconfig.Variable {
	return tfconfig.SetVariable(
		collections.Map(values, func(value string) tfconfig.Variable { return tfconfig.StringVariable(value) })...,
	)
}
