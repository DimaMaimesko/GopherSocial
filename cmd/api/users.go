package main

import (
	"context"
	"net/http"
	"strconv"

	"github.com/DimaMaimesko/GopherSocial/internal/store"
	"github.com/go-chi/chi/v5"
)

type userKey string

const userCtx userKey = "user"

// getUserHandler godoc
//
//	@Summary		Get user
//	@Description	Returns a user by ID.
//	@Tags			users
//	@Produce		json
//	@Param			userID	path		int	true	"User ID"
//	@Success		200		{object}	store.User
//	@Failure		404		{object}	map[string]string
//	@Failure		500		{object}	map[string]string
//	@Router			/users/{userID}/ [get]
func (app *application) getUserHandler(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)

	if err := app.jsonResponse(w, http.StatusOK, user); err != nil {
		app.internalServerError(w, r, err)
	}
}

type FollowerUser struct {
	UserID int64 `json:"user_id"`
}

// followUserHandler godoc
//
//	@Summary		Follow user
//	@Description	Follows a user by ID.
//	@Tags			users
//	@Produce		json
//	@Param			userID	path	int	true	"User ID"
//	@Success		204
//	@Failure		409	{object}	map[string]string
//	@Failure		500	{object}	map[string]string
//	@Router			/users/{userID}/follow [put]
func (app *application) followUserHandler(w http.ResponseWriter, r *http.Request) {
	userToFollow := getUserFromContext(r)

	// TODO: Revert back to auth userID from ctx
	followerUser := FollowerUser{
		UserID: 50, // Hardcoded follower ID
	}

	ctx := r.Context()

	if err := app.store.Followers.Follow(ctx, followerUser.UserID, userToFollow.ID); err != nil {
		switch err {
		case store.ErrNotFound:
			app.conflictResponse(w, r, err)
			return
		default:
			app.internalServerError(w, r, err)
			return
		}
	}

	if err := app.jsonResponse(w, http.StatusNoContent, nil); err != nil {
		app.internalServerError(w, r, err)
	}
}

// unfollowUserHandler godoc
//
//	@Summary		Unfollow user
//	@Description	Unfollows a user by ID.
//	@Tags			users
//	@Produce		json
//	@Param			userID	path	int	true	"User ID"
//	@Success		204
//	@Failure		500	{object}	map[string]string
//	@Router			/users/{userID}/unfollow [put]
func (app *application) unfollowUserHandler(w http.ResponseWriter, r *http.Request) {
	userToUnfollow := getUserFromContext(r)

	// TODO: Revert back to auth userID from ctx
	followerUser := FollowerUser{
		UserID: 50, // Hardcoded follower ID
	}

	ctx := r.Context()

	if err := app.store.Followers.Unfollow(ctx, followerUser.UserID, userToUnfollow.ID); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusNoContent, nil); err != nil {
		app.internalServerError(w, r, err)
	}
}

func (app *application) userContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
		if err != nil {
			app.badRequestResponse(w, r, err)
			return
		}

		ctx := r.Context()

		user, err := app.store.Users.GetByID(ctx, userID)
		if err != nil {
			switch err {
			case store.ErrNotFound:
				app.notFoundResponse(w, r, err)
				return
			default:
				app.internalServerError(w, r, err)
				return
			}
		}

		ctx = context.WithValue(ctx, userCtx, user)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getUserFromContext(r *http.Request) *store.User {
	user, _ := r.Context().Value(userCtx).(*store.User)
	return user
}
