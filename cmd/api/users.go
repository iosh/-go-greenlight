package main

import (
	"net/http"

	"github.com/iosh/go-greenlight/internal/data"
	"github.com/iosh/go-greenlight/internal/validator"
)

func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {

	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &input)

	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user := &data.User{
		Name:      input.Name,
		Email:     input.Email,
		Activated: false,
	}

	if err := user.Password.Set(input.Password); err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	v := validator.New()

	if data.ValidateUser(v, user); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	if err := app.models.Users.Insert(user); err != nil {
		app.errorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	if err := app.writeJSON(w, http.StatusCreated, envelop{"user": user}); err != nil {

		app.errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

}
