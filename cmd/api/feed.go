package main

import (
	"net/http"

	"github.com/DimaMaimesko/GopherSocial/internal/store"
)

// getUserFeedHandler godoc
//
//	@Summary		Get user feed
//	@Description	Returns the authenticated user's feed.
//	@Tags			feed
//	@Produce		json
//	@Param			limit	query		int			false	"Limit"
//	@Param			offset	query		int			false	"Offset"
//	@Param			sort	query		string		false	"Sort direction"	Enums(asc, desc)
//	@Param			search	query		string		false	"Search term"
//	@Param			tags	query		[]string	false	"Tags"
//	@Success		200		{array}		store.PostWithMetadata
//	@Failure		400		{object}	map[string]string
//	@Failure		500		{object}	map[string]string
//	@Router			/users/feed [get]
func (app *application) getUserFeedHandler(w http.ResponseWriter, r *http.Request) {
	fq := store.PaginatedFeedQuery{
		Limit:  20,
		Offset: 0,
		Sort:   "desc",
	}

	fq, err := fq.Parse(r)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(fq); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	ctx := r.Context()

	feed, err := app.store.Posts.GetUserFeed(ctx, int64(50), fq)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, feed); err != nil {
		app.internalServerError(w, r, err)
	}
}
