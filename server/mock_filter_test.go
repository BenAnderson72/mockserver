package server

import (
	"io"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
	server_utils "mockserver/server/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMockHandlerFiltersFileBackedDataByUserHeader(t *testing.T) {
	handler := &MockHandler{
		status:       http.StatusOK,
		mockFileData: []byte(`[{"id":1,"name":"John"},{"id":2,"name":"Jane"}]`),
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
	req.Header.Set("X-User-ID", "2")
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
		map[string]string{"x-user-id": "2"},
	)

	assert.Equal(t, "1", params["id"])
}
