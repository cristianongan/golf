package controllers

import (
	"errors"
	"start/constants"
	"start/controllers/request"
	"start/datasources"
	"start/models"
	"start/utils"
	"start/utils/response_message"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CPlayerScore struct{}

func (_ *CPlayerScore) CreatePlayerScore(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	body := request.CreatePlayerScoreBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}
	now := utils.GetTimeNow()

	// Batch insert player score
	playerScore := models.PlayerScore{}

	var listPlayer []models.PlayerScore

	for _, player := range body.Players {
		playerScore := models.PlayerScore{
			PartnerUid:  body.PartnerUid,
			CourseUid:   body.CourseUid,
			BookingDate: body.BookingDate,
			FlightId:    body.FlightId,
			Bag:         player.Bag,
			Course:      body.Course,
			Hole:        body.Hole,
			HoleIndex:   body.HoleIndex,
			Par:         body.Par,
			Shots:       player.Shots,
			Index:       player.Index,
			TimeStart:   body.TimeStart,
			TimeEnd:     body.TimeEnd,
		}

		playerScore.ModelId.CreatedAt = now.Unix()
		playerScore.ModelId.UpdatedAt = now.Unix()
		if playerScore.ModelId.Status == "" {
			playerScore.ModelId.Status = constants.STATUS_ENABLE
		}

		if playerScore.IsDuplicated(db) {
			response_message.BadRequest(c, constants.API_ERR_DUPLICATED_RECORD)
			return
		}

		listPlayer = append(listPlayer, playerScore)
	}

	errC := playerScore.BatchInsert(db, listPlayer)
	if errC != nil {
		response_message.InternalServerError(c, errC.Error())
		return
	}

	res := map[string]interface{}{
		"data": listPlayer,
	}

	okResponse(c, res)
}

func (_ *CPlayerScore) GetListPlayerScore(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	form := request.GetListPlayerScoreForm{}
	if bindErr := c.ShouldBind(&form); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	page := models.Page{
		Limit: form.PageRequest.Limit,
		Page:  form.PageRequest.Page,
		// SortBy:  form.PageRequest.SortBy,
		// SortDir: form.PageRequest.SortDir,
	}

	PlayerScoreR := models.PlayerScore{
		PartnerUid:  form.PartnerUid,
		CourseUid:   form.CourseUid,
		BookingDate: form.BookingDate,
		Bag:         form.Bag,
		Hole:        form.Hole,
		HoleIndex:   form.HoleIndex,
		FlightId:    form.FlightId,
	}
	list, total, err := PlayerScoreR.FindList(db, page, form.Status)
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	res := map[string]interface{}{
		"total": total,
		"data":  list,
	}

	okResponse(c, res)
}

func (_ *CPlayerScore) UpdatePlayerScore(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	playerScoreIdStr := c.Param("id")
	playerScoreId, err := strconv.ParseInt(playerScoreIdStr, 10, 64)
	if err != nil || playerScoreId <= 0 {
		response_message.BadRequest(c, errors.New("Id not valid").Error())
		return
	}

	// validate body
	body := request.UpdatePlayerScoreBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	playerScore := models.PlayerScore{}
	playerScore.Id = playerScoreId
	errF := playerScore.FindFirst(db)
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	if body.Course != "" {
		playerScore.Course = body.Course
	}

	playerScore.Hole = body.Hole
	playerScore.Par = body.Par
	playerScore.Shots = body.Shots
	playerScore.Index = body.Index
	playerScore.TimeStart = body.TimeStart
	playerScore.TimeEnd = body.TimeEnd

	errUdp := playerScore.Update(db)
	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	okResponse(c, playerScore)
}

func (_ *CPlayerScore) UpdateListPlayerScore(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	// validate body
	body := request.UpdateListPSBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	for _, player := range body.ListPlayer {
		playerScore := models.PlayerScore{}
		playerScore.Id = player.Id
		errF := playerScore.FindFirst(db)
		if errF != nil {
			response_message.InternalServerError(c, errF.Error())
			return
		}

		if player.Course != "" {
			playerScore.Course = player.Course
		}

		playerScore.Shots = player.Shots
		playerScore.TimeEnd = player.TimeEnd

		errUdp := playerScore.Update(db)
		if errUdp != nil {
			response_message.InternalServerError(c, errUdp.Error())
			return
		}
	}

	okRes(c)
}

func (_ *CPlayerScore) DeletePlayerScore(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)

	PlayerScoreIdStr := c.Param("id")
	PlayerScoreId, err := strconv.ParseInt(PlayerScoreIdStr, 10, 64)
	if err != nil || PlayerScoreId <= 0 {
		response_message.BadRequest(c, errors.New("Id not valid").Error())
		return
	}

	playerScore := models.PlayerScore{}
	playerScore.Id = PlayerScoreId
	errF := playerScore.FindFirst(db)
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	errDel := playerScore.Delete(db)
	if errDel != nil {
		response_message.InternalServerError(c, errDel.Error())
		return
	}

	okRes(c)
}
