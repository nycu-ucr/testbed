package router

import (
	// "net/http"
	"testbed/http/http_server/context"
	"testbed/http/http_server/logger"
	"testbed/http/http_server/producer"
	"testbed/httpwrapper"

	"github.com/nycu-ucr/gonet/http"

	"github.com/nycu-ucr/gin"
	// "github.com/gin-gonic/gin"
)

// Find all users
func GetUser(c *gin.Context) {
	// logger.ServerLog.Info("Handle GetUser")

	req := httpwrapper.NewRequest(c.Request, nil)
	if id, exists := c.Params.Get("id"); exists {
		req.Params["id"] = id
	}

	rsp := producer.HandleGetUser(req)
	c.JSON(http.StatusOK, rsp)
}

// Create a new user
func PostUser(c *gin.Context) {
	// logger.ServerLog.Info("Handle PostUser")

	user := context.User{}
	err := c.ShouldBindJSON(&user)

	if err != nil {
		logger.ServerLog.Errorf("PostUser Error: %+v", err)
		c.JSON(http.StatusNotAcceptable, "Error : "+err.Error())
		return
	} else {
		rsp := producer.HandlePostUser(&user)
		c.JSON(http.StatusOK, rsp)
	}
}

// Update the user information
func PutUser(c *gin.Context) {
	// logger.ServerLog.Info("Handle PutUser")

	user := context.User{}
	err := c.ShouldBindJSON(&user)

	if err != nil {
		logger.ServerLog.Errorf("Error: %+v", err)
		c.JSON(http.StatusNotAcceptable, "Error : "+err.Error())
		return
	} else {
		rsp := producer.HandlePutUser(&user)
		c.JSON(http.StatusOK, rsp)
	}
}

// Delete the user
func DeleteUser(c *gin.Context) {
	// logger.ServerLog.Info("Handle DeleteUser")

	req := httpwrapper.NewRequest(c.Request, nil)
	if id, exists := c.Params.Get("id"); exists {
		req.Params["id"] = id
	}

	rsp := producer.HandleDeleteUser(req)
	c.JSON(http.StatusOK, rsp)
}
