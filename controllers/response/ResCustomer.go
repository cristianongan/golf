package response

import (
	"start/controllers/request"
)

type CustomerRes struct {
	Status  string               `json:"status"`
	Message string               `json:"message"`
	Item    request.CustomerBody `json:"item"`
}
