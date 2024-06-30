package app

import (
	"context"
	"net/http"
)

func (c *App) PingHomeserver() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		type response struct {
			TransactionID string `json:"transaction_id"`
		}

		p, err := ReadRequestJSON(r, w, &response{})

		if err != nil {
			RespondWithError(w, &JSONResponse{
				Code: http.StatusOK,
				JSON: map[string]any{
					"errcode": "M_BAD_JSON",
					"error":   "Content not JSON.",
				},
			})
			return
		}

		c.Log.Info().Msgf("Transaction: %v", p.TransactionID)

		RespondWithJSON(w, &JSONResponse{
			Code: http.StatusOK,
		})

	}
}

func (c *App) PingHomeserverl(txnID string) error {

	resp, err := c.Matrix.AppservicePing(context.Background(), c.Config.AppService.ID, txnID)

	if err != nil {
		return err
	}

	c.Log.Info().Msgf("Ping response: %v", resp)

	return nil
}

func (c *App) RespondToPing() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		type response struct {
			TransactionID string `json:"transaction_id"`
		}

		ping, err := ReadRequestJSON(r, w, &response{})

		if err != nil {
			RespondWithError(w, &JSONResponse{
				Code: http.StatusInternalServerError,
				JSON: map[string]any{
					"errcode": "M_BAD_JSON",
					"error":   "Content not JSON.",
				},
			})
			return
		}

		c.Log.Info().Msgf("Transaction ID: %v", ping.TransactionID)

		RespondWithJSON(w, &JSONResponse{
			Code: http.StatusOK,
		})

	}
}
