package utils

import "github.com/hashicorp/terraform-plugin-framework/types"

func AsInt(value types.Int64) int {
	return int(value.ValueInt64())
}
