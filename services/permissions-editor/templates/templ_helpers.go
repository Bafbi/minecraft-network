// services/permissions-editor/templates/templ_helpers.go
package templates

import (
	"fmt"
	"strings"

	authpb "github.com/bafbi/minecraft-network/services/permissions-checker/auth"
)

func GetPolicyFieldString(policy *authpb.PolicyRule, fieldName string) string {
	if policy == nil {
		return ""
	}
	switch fieldName {
	case "id":
		return policy.GetId()
	case "targetAction":
		return policy.GetTargetAction()
	case "targetResource":
		return policy.GetTargetResource()
	case "playerConditionExpression":
		return policy.GetPlayerConditionExpression()
	case "serverConditionExpression":
		return policy.GetServerConditionExpression()
	case "effect":
		return policy.GetEffect()
	default:
		return ""
	}
}

func GetPolicyPriorityString(policy *authpb.PolicyRule) string {
	if policy != nil {
		return fmt.Sprintf("%d", policy.GetPriority())
	}
	return "100"
}

func GetPolicyConditionDefault(policy *authpb.PolicyRule, fieldName string) string {
	val := GetPolicyFieldString(policy, fieldName)
	if val == "" {
		return "true"
	}
	return val
}

func IsSelected(currentEffect, optionValue string) bool {
	return currentEffect == optionValue
}

func FormatMetadataForTextarea(metadata map[string]interface{}) string {
	var sb strings.Builder
	for k, v := range metadata {
		switch val := v.(type) {
		case string:
			sb.WriteString(fmt.Sprintf("%s=%s\n", k, val))
		case bool:
			sb.WriteString(fmt.Sprintf("%s=%t\n", k, val))
		case float64:
			if float64(int(val)) == val {
				sb.WriteString(fmt.Sprintf("%s=%d\n", k, int(val)))
			} else {
				sb.WriteString(fmt.Sprintf("%s=%f\n", k, val))
			}
		case []interface{}:
			var s []string
			for _, item := range val {
				s = append(s, fmt.Sprintf("%v", item))
			}
			sb.WriteString(fmt.Sprintf("%s=%s\n", k, strings.Join(s, ",")))
		default:
			sb.WriteString(fmt.Sprintf("%s=%v\n", k, val))
		}
	}
	return sb.String()
}

func GetPolicyFormButtonText(policy *authpb.PolicyRule) string {
	if policy != nil {
		return "Update Policy"
	}
	return "Add Policy"
}
