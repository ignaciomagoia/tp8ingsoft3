package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRegisterAndLoginFlow(t *testing.T) {
	app := newTestApp()

	registerPayload := map[string]string{
		"email":    "User@example.com",
		"password": "secret",
	}
	registerBody, err := json.Marshal(registerPayload)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(registerBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	app.router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code)

	var registerResp map[string]string
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &registerResp))
	require.NotEmpty(t, registerResp["message"])

	loginPayload := map[string]string{
		"email":    "user@example.com",
		"password": "secret",
	}
	loginBody, err := json.Marshal(loginPayload)
	require.NoError(t, err)

	loginReq := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(loginBody))
	loginReq.Header.Set("Content-Type", "application/json")
	loginRec := httptest.NewRecorder()
	app.router.ServeHTTP(loginRec, loginReq)

	require.Equal(t, http.StatusOK, loginRec.Code)

	var loginResp map[string]string
	require.NoError(t, json.Unmarshal(loginRec.Body.Bytes(), &loginResp))
	require.Equal(t, "login exitoso", loginResp["message"])

	listReq := httptest.NewRequest(http.MethodGet, "/users", nil)
	listRec := httptest.NewRecorder()
	app.router.ServeHTTP(listRec, listReq)

	require.Equal(t, http.StatusOK, listRec.Code)

	var listResp struct {
		Users []map[string]string `json:"users"`
	}
	require.NoError(t, json.Unmarshal(listRec.Body.Bytes(), &listResp))
	require.Len(t, listResp.Users, 1)
	require.NotContains(t, listResp.Users[0], "password")
	require.Equal(t, "user@example.com", listResp.Users[0]["email"])
}

func TestRegisterRejectsDuplicates(t *testing.T) {
	app := newTestApp()

	payload := map[string]string{
		"email":    "duplicate@example.com",
		"password": "secret",
	}
	body, err := json.Marshal(payload)
	require.NoError(t, err)

	req1 := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
	req1.Header.Set("Content-Type", "application/json")
	rec1 := httptest.NewRecorder()
	app.router.ServeHTTP(rec1, req1)

	require.Equal(t, http.StatusCreated, rec1.Code)

	req2 := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
	req2.Header.Set("Content-Type", "application/json")
	rec2 := httptest.NewRecorder()
	app.router.ServeHTTP(rec2, req2)

	require.Equal(t, http.StatusConflict, rec2.Code)
}

func TestLoginInvalidCredentials(t *testing.T) {
	app := newTestApp()

	payload := map[string]string{
		"email":    "unknown@example.com",
		"password": "secret",
	}
	body, err := json.Marshal(payload)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	app.router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestClearUsersEndpoint(t *testing.T) {
	app := newTestApp()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/users", nil)
	app.router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
}
