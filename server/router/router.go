package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Route is the information for every URI.
type Route struct {
	// Name is the name of this Route.
	Name string
	// Method is the string for the HTTP method. ex) GET, POST etc..
	Method string
	// Pattern is the pattern of the URI.
	Pattern string
	// HandlerFunc is the handler function of this route.
	HandlerFunc gin.HandlerFunc
}

// Routes is the list of the generated Route.
type Routes []Route

func NewGin() *gin.Engine {
	engine := gin.New()
	return engine
}

func AddService(engine *gin.Engine) *gin.RouterGroup {
	group := engine.Group("/test-server")

	for _, route := range routes {
		switch route.Method {
		case "GET":
			group.GET(route.Pattern, route.HandlerFunc)
		case "POST":
			group.POST(route.Pattern, route.HandlerFunc)
		case "PUT":
			group.PUT(route.Pattern, route.HandlerFunc)
		case "DELETE":
			group.DELETE(route.Pattern, route.HandlerFunc)
		}
	}

	return group
}

// Index is the index handler.
func Index(c *gin.Context) {
	c.String(http.StatusOK, "Hello World!")
}

var routes = Routes{
	{
		"Index",
		"GET",
		"/",
		Index,
	},
	{
		"Get All User",
		"GET",
		"/GetUser",
		GetUser,
	},
	{
		"Get The User",
		"GET",
		"/GetUser/:id",
		GetUser,
	},
	{
		"Post User",
		"POST",
		"/PostUser",
		PostUser,
	},
	{
		"Put User",
		"PUT",
		"/PutUser",
		PutUser,
	},
	{
		"Delete User",
		"DELETE",
		"/DeleteUser/:id",
		DeleteUser,
	},
}
