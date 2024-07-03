package utils

import (
	"github.com/labd/terraform-provider-folge/internal/folge"
)

func GetClient(data any) folge.ClientWithResponsesInterface {
	c, ok := data.(folge.ClientWithResponsesInterface)
	if !ok {
		panic("invalid client type")
	}
	return c
}
