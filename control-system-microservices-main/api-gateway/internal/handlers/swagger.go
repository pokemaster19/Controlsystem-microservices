package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/SpiritFoxo/control-system-microservices/api-gateway/internal/config"
	"github.com/gin-gonic/gin"
)

func SwaggerHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		usersSwagger, err := fetchSwagger(cfg.UsersServiceURL + "/swagger/doc.json")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users swagger: " + err.Error()})
			return
		}

		ordersSwagger, err := fetchSwagger(cfg.OrdersServiceURL + "/swagger/doc.json")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch orders swagger: " + err.Error()})
			return
		}

		mergedPaths := mergePaths(
			removeLocalSecurity(usersSwagger["paths"]),
			removeLocalSecurity(ordersSwagger["paths"]),
		)

		mergedSchemas := mergeSchemas(
			safeGetSchemas(usersSwagger),
			safeGetSchemas(ordersSwagger),
		)

		updateRefs(mergedPaths)
		updateRefs(mergedSchemas)

		mergedSwagger := map[string]interface{}{
			"openapi": "3.0.0",
			"info": map[string]interface{}{
				"title":   "Control System API",
				"version": "1.0",
			},
			"servers": []map[string]interface{}{
				{"url": "/api/v1", "description": "API Gateway"},
			},
			"paths": mergedPaths,
			"components": map[string]interface{}{
				"schemas": mergedSchemas,
				"securitySchemes": map[string]interface{}{
					"BearerAuth": map[string]interface{}{
						"type":         "http",
						"scheme":       "bearer",
						"bearerFormat": "JWT",
					},
				},
			},
			"security": []map[string][]string{
				{"BearerAuth": {}},
			},
		}

		c.JSON(http.StatusOK, mergedSwagger)
	}
}

func fetchSwagger(url string) (map[string]interface{}, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status: %v", resp.StatusCode)
	}

	var swagger map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&swagger); err != nil {
		return nil, err
	}
	return swagger, nil
}

func mergePaths(a, b interface{}) map[string]interface{} {
	merged := make(map[string]interface{})
	if am, ok := a.(map[string]interface{}); ok {
		for k, v := range am {
			merged[k] = v
		}
	}
	if bm, ok := b.(map[string]interface{}); ok {
		for k, v := range bm {
			merged[k] = v
		}
	}
	return merged
}

func mergeSchemas(a, b interface{}) map[string]interface{} {
	merged := make(map[string]interface{})
	if am, ok := a.(map[string]interface{}); ok {
		for k, v := range am {
			merged[k] = v
		}
	}
	if bm, ok := b.(map[string]interface{}); ok {
		for k, v := range bm {
			merged[k] = v
		}
	}
	return merged
}

func removeLocalSecurity(paths interface{}) map[string]interface{} {
	if pm, ok := paths.(map[string]interface{}); ok {
		for path, ops := range pm {
			if methods, ok := ops.(map[string]interface{}); ok {
				for method, details := range methods {
					if d, ok := details.(map[string]interface{}); ok {
						delete(d, "security")
						methods[method] = d
					}
				}
			}
			pm[path] = ops
		}
		return pm
	}
	return nil
}

func safeGetSchemas(swagger map[string]interface{}) interface{} {
	if swagger == nil {
		return nil
	}
	if comp, ok := swagger["components"].(map[string]interface{}); ok {
		if schemas, ok := comp["schemas"]; ok {
			return schemas
		}
	}
	if definitions, ok := swagger["definitions"]; ok {
		return definitions
	}
	return nil
}

func updateRefs(obj interface{}) {
	switch v := obj.(type) {
	case map[string]interface{}:
		for key, val := range v {
			if key == "$ref" {
				if str, ok := val.(string); ok {
					v[key] = strings.Replace(str, "#/definitions/", "#/components/schemas/", -1)
				}
			} else {
				updateRefs(val)
			}
		}
	case []interface{}:
		for _, item := range v {
			updateRefs(item)
		}
	}
}
