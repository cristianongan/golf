package middlewares

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"regexp"
	"start/config"
	"start/constants"
	"start/models"
	"start/utils"
	"time"

	"github.com/gin-gonic/gin"
)

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

type logmsg struct {
	Environment string    `json:"environment"`
	Path        string    `json:"path"`
	TimeStamp   time.Time `json:"@timestamp"`
	LocalTime   time.Time `json:"local_time"`
	CreatedAt   int64     `json:"created_at"`
	RawPath     string    `json:"rawpath"`
	Header      string    `json:"header"`
	Module      string    `json:"module"`
	Method      string    `json:"method"`
	ClientIP    string    `json:"ip"`
	Duration    int64     `json:"duration"`
	Request     string    `json:"request"`
	PostForm    string    `json:"postform"`
	Response    string    `json:"response"`
	Host        string    `json:"host"`
	Location    string    `json:"location"`
	Status      int       `json:"status"`
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func GinBodyLogMiddleware(c *gin.Context) {
	newlogmessage := logmsg{}
	start := utils.GetTimeNow().UTC()
	newlogmessage.TimeStamp = utils.GetTimeNow()
	newlogmessage.LocalTime = utils.GetLocalUnixTime()
	newlogmessage.CreatedAt = utils.TimeStampMilisecond(newlogmessage.TimeStamp.UnixNano())
	buf, _ := ioutil.ReadAll(c.Request.Body)
	rdr1 := ioutil.NopCloser(bytes.NewBuffer(buf))
	rdr2 := ioutil.NopCloser(bytes.NewBuffer(buf)) //We have to create a new Buffer, because rdr1 will be read.

	c.Request.Body = rdr2
	newlogmessage.Request = readBody(rdr1)
	blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
	c.Writer = blw
	c.Next()
	end := utils.GetTimeNow().UTC()
	newlogmessage.Environment = config.GetEnvironmentName()
	newlogmessage.Module = config.GetModuleName()
	newlogmessage.Path = c.Request.URL.Path
	newlogmessage.Status = c.Writer.Status()
	newlogmessage.RawPath = c.Request.URL.RawQuery
	newlogmessage.Header = ""
	for k, v := range c.Request.Header {
		newlogmessage.Header += fmt.Sprintf("%q:%q;", k, v)
	}
	newlogmessage.PostForm = ""
	c.Request.ParseForm()
	for k, v := range c.Request.PostForm {
		newlogmessage.PostForm += fmt.Sprintf("%q:%q;", k, v)
	}
	latency := end.Sub(start)
	latencyInMilliseconds := int64(latency / time.Millisecond)
	newlogmessage.Method = c.Request.Method
	newlogmessage.ClientIP = c.ClientIP()
	newlogmessage.Duration = latencyInMilliseconds
	newlogmessage.Response = blw.body.String()
	newlogmessage.Host = ""
	host, _ := os.Hostname()
	addrs, _ := net.LookupIP(host)
	for _, addr := range addrs {
		if ipv4 := addr.To4(); ipv4 != nil {
			newlogmessage.Host = ipv4.String()
		}
	}

	// user name
	userName := ""
	value, exists := c.Get(constants.CMS_USER_PROFILE_KEY)
	if exists {
		baseInfo, isUserProfile := value.(models.CmsUserProfile)
		if isUserProfile {
			userName = baseInfo.UserName
		}
	}

	if newlogmessage.Method != "OPTIONS" {
		newlog := "[REQ_LOG] " + utils.NumberToString(newlogmessage.LocalTime) +
			" |user_name:" + userName +
			" |status:" + utils.NumberToString(newlogmessage.Status) +
			" |duration:" + utils.NumberToString(newlogmessage.Duration) +
			" |client_ip:" + newlogmessage.ClientIP +
			" |method:" + newlogmessage.Method +
			" |path:" + newlogmessage.Path +
			" |request:" + newlogmessage.Request +
			" |response:" + newlogmessage.Response

		notShowrequests := []string{
			"/golf-cms/metrics",
		}
		for _, item := range notShowrequests {
			matched, _ := regexp.MatchString(item, c.Request.URL.Path)
			if !matched {
				fmt.Println(newlog)
			}
		}
	}

}

func readBody(reader io.Reader) string {
	buf := new(bytes.Buffer)
	buf.ReadFrom(reader)

	s := buf.String()
	return s
}
