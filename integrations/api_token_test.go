// Copyright 2018 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package integrations

import (
	"net/http"
	"testing"

	"code.gitea.io/gitea/models"
	"code.gitea.io/gitea/models/db"
	api "code.gitea.io/gitea/modules/structs"
)

// TestAPICreateAndDeleteToken tests that token that was just created can be deleted
func TestAPICreateAndDeleteToken(t *testing.T) {
	defer prepareTestEnv(t)()
	user := db.AssertExistsAndLoadBean(t, &models.User{ID: 1}).(*models.User)

	req := NewRequestWithJSON(t, "POST", "/api/v1/users/user1/tokens", map[string]string{
		"name": "test-key-1",
	})
	req = AddBasicAuthHeader(req, user.Name)
	resp := MakeRequest(t, req, http.StatusCreated)

	var newAccessToken api.AccessToken
	DecodeJSON(t, resp, &newAccessToken)
	db.AssertExistsAndLoadBean(t, &models.AccessToken{
		ID:    newAccessToken.ID,
		Name:  newAccessToken.Name,
		Token: newAccessToken.Token,
		UID:   user.ID,
	})

	req = NewRequestf(t, "DELETE", "/api/v1/users/user1/tokens/%d", newAccessToken.ID)
	req = AddBasicAuthHeader(req, user.Name)
	MakeRequest(t, req, http.StatusNoContent)

	db.AssertNotExistsBean(t, &models.AccessToken{ID: newAccessToken.ID})

	req = NewRequestWithJSON(t, "POST", "/api/v1/users/user1/tokens", map[string]string{
		"name": "test-key-2",
	})
	req = AddBasicAuthHeader(req, user.Name)
	resp = MakeRequest(t, req, http.StatusCreated)
	DecodeJSON(t, resp, &newAccessToken)

	req = NewRequestf(t, "DELETE", "/api/v1/users/user1/tokens/%s", newAccessToken.Name)
	req = AddBasicAuthHeader(req, user.Name)
	MakeRequest(t, req, http.StatusNoContent)

	db.AssertNotExistsBean(t, &models.AccessToken{ID: newAccessToken.ID})
}

// TestAPIDeleteMissingToken ensures that error is thrown when token not found
func TestAPIDeleteMissingToken(t *testing.T) {
	defer prepareTestEnv(t)()
	user := db.AssertExistsAndLoadBean(t, &models.User{ID: 1}).(*models.User)

	req := NewRequestf(t, "DELETE", "/api/v1/users/user1/tokens/%d", db.NonexistentID)
	req = AddBasicAuthHeader(req, user.Name)
	MakeRequest(t, req, http.StatusNotFound)
}
