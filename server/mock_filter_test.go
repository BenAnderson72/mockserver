package server

import (
	"io"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
	msconfig "mockserver/config"
	server_utils "mockserver/server/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMockHandlerFiltersFileBackedDataByConfiguredHeader(t *testing.T) {
	handler := &MockHandler{
		status:       http.StatusOK,
		mockFileData: []byte(`[{"id":1,"name":"John"},{"id":2,"name":"Jane"}]`),
		routecfg: msconfig.RouteConfig{
			FilterHeaders: map[string]string{"X-Account-ID": "id"},
		},
	}
	app := fiber.New()
	app.Get("/api/v1/users", func(c *fiber.Ctx) error {
		ctx := server_utils.EContext{
			Headers: buildHeaders(c),
			Query:   buildQuery(c),
			Path:    c.AllParams(),
			Body:    map[string]interface{}{},
		}
		return handler.handler(c, ctx)
	})

	req, err := http.NewRequest(http.MethodGet, "/api/v1/users", nil)
	require.NoError(t, err)
	req.Header.Set("X-Account-ID", "2")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.JSONEq(t, `[{"id":2,"name":"Jane"}]`, string(body))
}

func TestBuildMockFilterParamsPrefersExistingQueryFilter(t *testing.T) {
	params := buildMockFilterParams(
		map[string]string{"id": "1"},
		map[string]string{"x-account-id": "2"},
		map[string]string{"X-Account-ID": "id"},
	)

	assert.Equal(t, "1", params["id"])
}

func TestBuildMockFilterParamsIgnoresUnconfiguredHeaders(t *testing.T) {
	params := buildMockFilterParams(
		map[string]string{},
		map[string]string{"x-user-id": "2"},
		nil,
	)

	assert.Empty(t, params)
}
