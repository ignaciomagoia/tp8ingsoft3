package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateListUpdateDeleteTodoFlow(t *testing.T) {
	app := newTestApp()

	// register a user to associate todos
	registerBody, err := json.Marshal(map[string]string{
		"email":    "tasks@example.com",
		"password": "secret",
	})
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(registerBody))
	req.Header.Set("Content-Type", "application/json")
	app.router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)

	// create todo
	createBody, err := json.Marshal(map[string]string{
		"email": "tasks@example.com",
		"title": "Primera tarea",
	})
	require.NoError(t, err)

	createRec := httptest.NewRecorder()
	createReq := httptest.NewRequest(http.MethodPost, "/todos", bytes.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	app.router.ServeHTTP(createRec, createReq)
	require.Equal(t, http.StatusCreated, createRec.Code)

	var createResp struct {
		Todo map[string]interface{} `json:"todo"`
	}
	require.NoError(t, json.Unmarshal(createRec.Body.Bytes(), &createResp))
	require.Equal(t, "Primera tarea", createResp.Todo["title"])
	require.Equal(t, false, createResp.Todo["completed"])
	todoID, ok := createResp.Todo["id"].(string)
	require.True(t, ok)

	// list todos filtered by email
	listRec := httptest.NewRecorder()
	listReq := httptest.NewRequest(http.MethodGet, "/todos?email=tasks@example.com", nil)
	app.router.ServeHTTP(listRec, listReq)
	require.Equal(t, http.StatusOK, listRec.Code)

	var listResp struct {
		Todos []map[string]interface{} `json:"todos"`
	}
	require.NoError(t, json.Unmarshal(listRec.Body.Bytes(), &listResp))
	require.Len(t, listResp.Todos, 1)
	require.Equal(t, todoID, listResp.Todos[0]["id"])

	// update todo
	updateBody, err := json.Marshal(map[string]interface{}{
		"title":     "Actualizada",
		"completed": true,
	})
	require.NoError(t, err)

	updateRec := httptest.NewRecorder()
	updateReq := httptest.NewRequest(http.MethodPut, "/todos/"+todoID, bytes.NewReader(updateBody))
	updateReq.Header.Set("Content-Type", "application/json")
	app.router.ServeHTTP(updateRec, updateReq)
	require.Equal(t, http.StatusOK, updateRec.Code)

	var updateResp struct {
		Todo map[string]interface{} `json:"todo"`
	}
	require.NoError(t, json.Unmarshal(updateRec.Body.Bytes(), &updateResp))
	require.Equal(t, "Actualizada", updateResp.Todo["title"])
	require.Equal(t, true, updateResp.Todo["completed"])

	// delete todo
	deleteRec := httptest.NewRecorder()
	deleteReq := httptest.NewRequest(http.MethodDelete, "/todos/"+todoID, nil)
	app.router.ServeHTTP(deleteRec, deleteReq)
	require.Equal(t, http.StatusOK, deleteRec.Code)

	// ensure list is empty after delete
	listRec2 := httptest.NewRecorder()
	listReq2 := httptest.NewRequest(http.MethodGet, "/todos?email=tasks@example.com", nil)
	app.router.ServeHTTP(listRec2, listReq2)
	require.Equal(t, http.StatusOK, listRec2.Code)
	require.NoError(t, json.Unmarshal(listRec2.Body.Bytes(), &listResp))
	require.Len(t, listResp.Todos, 0)
}

func TestTodoValidationErrors(t *testing.T) {
	app := newTestApp()

	// create with invalid payload
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/todos", bytes.NewReader([]byte(`{"email":"","title":""}`)))
	req.Header.Set("Content-Type", "application/json")
	app.router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code)

	// update with invalid ID
	updateRec := httptest.NewRecorder()
	updateReq := httptest.NewRequest(http.MethodPut, "/todos/invalid-id", bytes.NewReader([]byte(`{"completed":true}`)))
	updateReq.Header.Set("Content-Type", "application/json")
	app.router.ServeHTTP(updateRec, updateReq)
	require.Equal(t, http.StatusBadRequest, updateRec.Code)

	// delete with invalid ID
	deleteRec := httptest.NewRecorder()
	deleteReq := httptest.NewRequest(http.MethodDelete, "/todos/invalid-id", nil)
	app.router.ServeHTTP(deleteRec, deleteReq)
	require.Equal(t, http.StatusBadRequest, deleteRec.Code)
}

func TestClearTodosEndpoint(t *testing.T) {
	app := newTestApp()

	registerBody, err := json.Marshal(map[string]string{
		"email":    "tasks@example.com",
		"password": "secret",
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(registerBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	app.router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)

	createBody, err := json.Marshal(map[string]string{
		"email": "tasks@example.com",
		"title": "Primera tarea",
	})
	require.NoError(t, err)

	createReq := httptest.NewRequest(http.MethodPost, "/todos", bytes.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createRec := httptest.NewRecorder()
	app.router.ServeHTTP(createRec, createReq)
	require.Equal(t, http.StatusCreated, createRec.Code)

	clearReq := httptest.NewRequest(http.MethodDelete, "/todos?email=tasks@example.com", nil)
	clearRec := httptest.NewRecorder()
	app.router.ServeHTTP(clearRec, clearReq)
	require.Equal(t, http.StatusOK, clearRec.Code)
}
