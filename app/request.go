package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type JSONResponse struct {
	Code    int
	JSON    interface{}
	Headers map[string]string
}

func ReadRequestJSON[T any](r *http.Request, w http.ResponseWriter, p T) (T, error) {

	type Response struct {
		Error string
	}

	if r.Body == nil {
		return p, errors.New("Request body is empty.")
	}

	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		return p, errors.New("Bad request.")
	}

	return p, nil
}

func MessageResponse(code int, msg string) *JSONResponse {
	return &JSONResponse{
		Code: code,
		JSON: struct {
			Message string `json:"message"`
		}{msg},
	}
}

func RespondWithJSON(w http.ResponseWriter, res *JSONResponse) {
	response, err := json.Marshal(res.JSON)
	if err != nil {
		res = MessageResponse(500, "Internal Server Error")
		response, _ = json.Marshal(res.JSON)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(res.Code)
	w.Write(response)
}

func RespondWithError(w http.ResponseWriter, res *JSONResponse) {
	response, err := json.Marshal(res.JSON)
	if err != nil {
		res = MessageResponse(500, "Internal Server Error")
		response, _ = json.Marshal(res.JSON)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(res.Code)
	w.Write(response)
}

func RespondWithBadRequestError(w http.ResponseWriter) {
	RespondWithJSON(w, &JSONResponse{
		Code: http.StatusOK,
		JSON: map[string]any{
			"error": "bad request",
		},
	})
}

func (c *App) RobotsTXT() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintln(w, "User-agent: *")
		fmt.Fprintln(w, "Disallow: /")
	}
}
