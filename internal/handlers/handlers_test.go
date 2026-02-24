package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sc23bd/COMP3011_Coursework1/internal/db/memory"
	"github.com/sc23bd/COMP3011_Coursework1/internal/handlers"
	"github.com/sc23bd/COMP3011_Coursework1/internal/models"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// newRouter builds a minimal Gin engine wired to a fresh store.
func newRouter() *gin.Engine {
	store := memory.NewStore()
	h := handlers.NewHandler(store)

	r := gin.New()
	v1 := r.Group("/api/v1")
	{
		items := v1.Group("/items")
		items.GET("", h.ListItems)
		items.POST("", h.CreateItem)
		items.GET("/:id", h.GetItem)
		items.PUT("/:id", h.UpdateItem)
		items.DELETE("/:id", h.DeleteItem)
	}
	return r
}

// doRequest executes an HTTP request against the router and returns the recorder.
func doRequest(r *gin.Engine, method, path string, body interface{}) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// --- ListItems ---------------------------------------------------------------

// TestListItems_Empty verifies that an empty store returns 200 with an empty
// data slice, satisfying the Uniform Interface principle (consistent structure).
func TestListItems_Empty(t *testing.T) {
	r := newRouter()
	w := doRequest(r, http.MethodGet, "/api/v1/items", nil)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp models.ItemsResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if resp.Data == nil {
		t.Fatal("expected non-nil data slice")
	}
	if len(resp.Data) != 0 {
		t.Fatalf("expected 0 items, got %d", len(resp.Data))
	}
}

// --- CreateItem --------------------------------------------------------------

// TestCreateItem_Success checks that a valid payload yields 201, a Location
// header, and HATEOAS links (Uniform Interface / HATEOAS principle).
func TestCreateItem_Success(t *testing.T) {
	r := newRouter()
	w := doRequest(r, http.MethodPost, "/api/v1/items", map[string]string{
		"name":        "Widget",
		"description": "A test widget",
	})

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	if w.Header().Get("Location") == "" {
		t.Fatal("expected Location header to be set")
	}

	var resp models.ItemResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if resp.ID == "" {
		t.Fatal("expected non-empty ID")
	}
	if resp.Name != "Widget" {
		t.Fatalf("expected name 'Widget', got %q", resp.Name)
	}
	if len(resp.Links) == 0 {
		t.Fatal("expected HATEOAS links")
	}
}

// TestCreateItem_MissingName ensures validation rejects a request with no name
// (Stateless principle — each request must be self-contained and valid).
func TestCreateItem_MissingName(t *testing.T) {
	r := newRouter()
	w := doRequest(r, http.MethodPost, "/api/v1/items", map[string]string{
		"description": "no name",
	})

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// --- GetItem -----------------------------------------------------------------

// TestGetItem_NotFound checks that a missing item returns 404.
func TestGetItem_NotFound(t *testing.T) {
	r := newRouter()
	w := doRequest(r, http.MethodGet, "/api/v1/items/999", nil)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

// TestGetItem_Success creates an item and then retrieves it by ID.
func TestGetItem_Success(t *testing.T) {
	r := newRouter()

	// Create
	w := doRequest(r, http.MethodPost, "/api/v1/items", map[string]string{
		"name": "Gadget",
	})
	var created models.ItemResponse
	_ = json.NewDecoder(w.Body).Decode(&created)

	// Fetch
	w = doRequest(r, http.MethodGet, "/api/v1/items/"+created.ID, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp models.ItemResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if resp.ID != created.ID {
		t.Fatalf("expected ID %s, got %s", created.ID, resp.ID)
	}
}

// --- UpdateItem --------------------------------------------------------------

// TestUpdateItem_Success verifies that PUT replaces the item representation.
func TestUpdateItem_Success(t *testing.T) {
	r := newRouter()

	// Create
	w := doRequest(r, http.MethodPost, "/api/v1/items", map[string]string{
		"name": "Original",
	})
	var created models.ItemResponse
	_ = json.NewDecoder(w.Body).Decode(&created)

	// Update
	w = doRequest(r, http.MethodPut, "/api/v1/items/"+created.ID, map[string]string{
		"name":        "Updated",
		"description": "now updated",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp models.ItemResponse
	_ = json.NewDecoder(w.Body).Decode(&resp)
	if resp.Name != "Updated" {
		t.Fatalf("expected name 'Updated', got %q", resp.Name)
	}
}

// TestUpdateItem_NotFound ensures updating a non-existent item returns 404.
func TestUpdateItem_NotFound(t *testing.T) {
	r := newRouter()
	w := doRequest(r, http.MethodPut, "/api/v1/items/999", map[string]string{
		"name": "Ghost",
	})

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

// --- DeleteItem --------------------------------------------------------------

// TestDeleteItem_Success creates an item and checks that DELETE removes it.
func TestDeleteItem_Success(t *testing.T) {
	r := newRouter()

	// Create
	w := doRequest(r, http.MethodPost, "/api/v1/items", map[string]string{
		"name": "Throwaway",
	})
	var created models.ItemResponse
	_ = json.NewDecoder(w.Body).Decode(&created)

	// Delete
	w = doRequest(r, http.MethodDelete, "/api/v1/items/"+created.ID, nil)
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}

	// Verify gone
	w = doRequest(r, http.MethodGet, "/api/v1/items/"+created.ID, nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 after delete, got %d", w.Code)
	}
}

// TestDeleteItem_NotFound ensures deleting a non-existent item returns 404.
func TestDeleteItem_NotFound(t *testing.T) {
	r := newRouter()
	w := doRequest(r, http.MethodDelete, "/api/v1/items/999", nil)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

// --- Stateless principle -----------------------------------------------------

// TestStateless_NoCookies verifies that the NoSessionState middleware rejects
// requests that carry a Cookie header.
func TestStateless_NoCookies(t *testing.T) {
	// Build router with NoSessionState middleware applied.
	store := memory.NewStore()
	h := handlers.NewHandler(store)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		if c.Request.Header.Get("Cookie") != "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "session cookies are not supported; the API is stateless",
			})
			return
		}
		c.Next()
	})
	v1 := r.Group("/api/v1")
	v1.GET("/items", h.ListItems)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/items", nil)
	req.Header.Set("Cookie", "session=abc123")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for cookie-bearing request, got %d", w.Code)
	}
}
