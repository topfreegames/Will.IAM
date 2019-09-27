package client

import (
	"fmt"
	"net/http"
	"strings"
)

func Generate(ownershipLevel, action string, resourceHierarchy ...string) func(r *http.Request) string {
	if len(resourceHierarchy) == 0 {
		resourceHierarchy = []string{"*"}
	}

	return func(r *http.Request) string {
		if wi := Get(r); wi != nil {
			return fmt.Sprintf(
				"%s::%s::%s::%s",
				wi.GetServiceName(),
				ownershipLevel,
				action,
				strings.Join(resourceHierarchy, "::"),
			)
		}

		return ""
	}
}

func GenerateInfo() func(r *http.Request) string {
	return func(r *http.Request) string {
		return ""
	}
}
