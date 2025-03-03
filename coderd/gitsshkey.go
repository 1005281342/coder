package coderd

import (
	"net/http"

	"github.com/coder/coder/coderd/database"
	"github.com/coder/coder/coderd/gitsshkey"
	"github.com/coder/coder/coderd/httpapi"
	"github.com/coder/coder/coderd/httpmw"
	"github.com/coder/coder/coderd/rbac"
	"github.com/coder/coder/codersdk"
)

func (api *API) regenerateGitSSHKey(rw http.ResponseWriter, r *http.Request) {
	user := httpmw.UserParam(r)

	if !api.Authorize(rw, r, rbac.ActionUpdate, rbac.ResourceUserData.WithOwner(user.ID.String())) {
		return
	}

	privateKey, publicKey, err := gitsshkey.Generate(api.SSHKeygenAlgorithm)
	if err != nil {
		httpapi.Write(rw, http.StatusInternalServerError, httpapi.Response{
			Message: "Internal error generating a new SSH keypair.",
			Detail:  err.Error(),
		})
		return
	}

	err = api.Database.UpdateGitSSHKey(r.Context(), database.UpdateGitSSHKeyParams{
		UserID:     user.ID,
		UpdatedAt:  database.Now(),
		PrivateKey: privateKey,
		PublicKey:  publicKey,
	})
	if err != nil {
		httpapi.Write(rw, http.StatusInternalServerError, httpapi.Response{
			Message: "Internal error updating user's git SSH key.",
			Detail:  err.Error(),
		})
		return
	}

	newKey, err := api.Database.GetGitSSHKey(r.Context(), user.ID)
	if err != nil {
		httpapi.Write(rw, http.StatusInternalServerError, httpapi.Response{
			Message: "Internal error fetching user's git SSH key.",
			Detail:  err.Error(),
		})
		return
	}

	httpapi.Write(rw, http.StatusOK, codersdk.GitSSHKey{
		UserID:    newKey.UserID,
		CreatedAt: newKey.CreatedAt,
		UpdatedAt: newKey.UpdatedAt,
		// No need to return the private key to the user
		PublicKey: newKey.PublicKey,
	})
}

func (api *API) gitSSHKey(rw http.ResponseWriter, r *http.Request) {
	user := httpmw.UserParam(r)

	if !api.Authorize(rw, r, rbac.ActionRead, rbac.ResourceUserData.WithOwner(user.ID.String())) {
		return
	}

	gitSSHKey, err := api.Database.GetGitSSHKey(r.Context(), user.ID)
	if err != nil {
		httpapi.Write(rw, http.StatusInternalServerError, httpapi.Response{
			Message: "Internal error fetching user's SSH key.",
			Detail:  err.Error(),
		})
		return
	}

	httpapi.Write(rw, http.StatusOK, codersdk.GitSSHKey{
		UserID:    gitSSHKey.UserID,
		CreatedAt: gitSSHKey.CreatedAt,
		UpdatedAt: gitSSHKey.UpdatedAt,
		// No need to return the private key to the user
		PublicKey: gitSSHKey.PublicKey,
	})
}

func (api *API) agentGitSSHKey(rw http.ResponseWriter, r *http.Request) {
	agent := httpmw.WorkspaceAgent(r)
	resource, err := api.Database.GetWorkspaceResourceByID(r.Context(), agent.ResourceID)
	if err != nil {
		httpapi.Write(rw, http.StatusInternalServerError, httpapi.Response{
			Message: "Internal error fetching workspace resource.",
			Detail:  err.Error(),
		})
		return
	}

	job, err := api.Database.GetWorkspaceBuildByJobID(r.Context(), resource.JobID)
	if err != nil {
		httpapi.Write(rw, http.StatusInternalServerError, httpapi.Response{
			Message: "Internal error fetching workspace build.",
			Detail:  err.Error(),
		})
		return
	}

	workspace, err := api.Database.GetWorkspaceByID(r.Context(), job.WorkspaceID)
	if err != nil {
		httpapi.Write(rw, http.StatusInternalServerError, httpapi.Response{
			Message: "Internal error fetching workspace.",
			Detail:  err.Error(),
		})
		return
	}

	gitSSHKey, err := api.Database.GetGitSSHKey(r.Context(), workspace.OwnerID)
	if err != nil {
		httpapi.Write(rw, http.StatusInternalServerError, httpapi.Response{
			Message: "Internal error fetching git SSH key.",
			Detail:  err.Error(),
		})
		return
	}

	httpapi.Write(rw, http.StatusOK, codersdk.AgentGitSSHKey{
		PublicKey:  gitSSHKey.PublicKey,
		PrivateKey: gitSSHKey.PrivateKey,
	})
}
