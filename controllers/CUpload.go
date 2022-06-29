package controllers

import (
	"log"
	"start/constants"
	"start/datasources"
	"start/utils/response_message"

	"github.com/gin-gonic/gin"
)

type CUpload struct{}

func (_ *CUpload) UploadImage(c *gin.Context) {
	//name := c.PostForm("name")
	//log.Println("name", name)

	type Sizer interface {
		Size() int64
	}
	file, _, errImg := c.Request.FormFile("image")
	if errImg != nil {
		response_message.BadRequest(c, errImg.Error())
		return
	}
	fileSize := file.(Sizer).Size()
	if fileSize > constants.MAX_SIZE_AVATAR_UPLOAD {
		response_message.BadRequest(c, "over limit size")
		return
	}

	// timeStamp := strconv.FormatInt(time.Now().Unix(), 10)
	// link, errUpdload := aws.UploadAvatar(&file, user.Model.Uid+"_"+timeStamp)
	link, errUpdload := datasources.UploadFile(&file)
	if errUpdload != nil {
		log.Println(errUpdload)
		response_message.InternalServerError(c, errUpdload.Error())
		return
	}

	res := map[string]interface{}{
		"link": link,
	}

	c.JSON(200, res)
}
