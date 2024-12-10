package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
)

func statusCheck(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK\n"))
}

func updateGames(w http.ResponseWriter, r *http.Request) {
	f, err := os.OpenFile(GAME_LIST_PATH, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		internalError(w, err)
		return
	}

	fi, err := f.Stat()
	if err != nil {
		internalError(w, err)
		return
	}

	gameList := map[string]string{}
	if fi.Size() > 0 {
		if err = json.NewDecoder(f).Decode(&gameList); err != nil {
			internalError(w, err)
			return
		}
	}

	games, err := tc.TopGames(r.Context())
	if err != nil {
		internalError(w, err)
		return
	}

	for _, g := range games {
		gameList[g.ID] = g.Title
	}

	b, err := json.MarshalIndent(gameList, "", "\t")
	if err != nil {
		internalError(w, err)
		return
	}

	_, err = f.WriteAt(b, 0)
	if err != nil {
		internalError(w, err)
		return
	}

	w.Write([]byte("ok"))
}

func internalError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(fmt.Sprintf("internal server error: %s\n", err.Error())))
}

func badRequest(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(fmt.Sprintf("bad request: %s\n", err.Error())))
}

func JoinError(errString string, errs ...error) error {
	return errors.Join(append(errs, errors.New(errString))...)
}

type BadRequest struct {
	Message string
}

func (br *BadRequest) Error() string {
	return "error: bad request " + br.Message
}

func processHandler(w http.ResponseWriter, r *http.Request) {
	processor := &Processor{}
	params, err := processor.ParseParams(r)
	if err != nil {
		badRequest(w, err)
		return
	}

	// title := GenerateTitle(params.t1, params.t2, params.GameID)
	title := "testing123"
	resultFile, err := os.Create("./" + title)
	if err != nil {
		internalError(w, JoinError("error: failed to create result file", err))
		return
	}

	result, err := process(r.Context(), params)
	if err != nil {
		internalError(w, err)
		return
	}

	if err := json.NewEncoder(resultFile).Encode(result); err != nil {
		internalError(w, errors.Join(errors.New("error: failed to write result file"), err))
		return
	}

	if err := json.NewEncoder(w).Encode(result.ProcessId); err != nil {
		internalError(w, errors.Join(errors.New("error: failed to encode response"), err))
		return
	}
}
