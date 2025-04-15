package gin

import (
	"github.com/gin-gonic/gin"
	"github.com/ph0m1/p_gateway/logging"
	"io/ioutil"
)

func DebugHandler(logger logging.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger.Debug("Method:", c.Request.Method)
		logger.Debug("URL:", c.Request.RequestURI)
		logger.Debug("Query:", c.Request.URL.Query())
		logger.Debug("Params:", c.Params)
		logger.Debug("Headers:", c.Request.Header)
		body, _ := ioutil.ReadAll(c.Request.Body)
		c.Request.Body.Close()
		logger.Debug("Body:", string(body))
		c.JSON(200, gin.H{
			"message": "pong",
		})
	}
}
