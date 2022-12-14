package producer

import (
	"net/http"
	"testbed/http/http_server/context"
	"testbed/http/http_server/logger"
	"testbed/httpwrapper"
)

func HandleGetUser(request *httpwrapper.Request) *httpwrapper.Response {
	id := request.Params["id"]

	userInfo, problemDetails := GetUserInformationProcedure(id)
	if problemDetails != nil {
		return httpwrapper.NewResponse(int(problemDetails.Status), nil, problemDetails)
	} else {
		return httpwrapper.NewResponse(http.StatusOK, nil, userInfo)
	}
}

func GetUserInformationProcedure(id string) (context.Users, *httpwrapper.ProblemDetails) {
	var users context.Users

	if id != "" {
		if user, ok := context.UserFindById(id); ok {
			userInfo := &context.User{
				Id:       user.Id,
				Name:     user.Name,
				Password: user.Password,
			}
			if &userInfo != nil {
				users = append(users, *userInfo)
			}
			logger.ServerLog.Infof("Find user: %+v", users)
		} else {
			logger.ServerLog.Warnf("User with Id=%s not found", id)
			problemDetails := &httpwrapper.ProblemDetails{
				Status: http.StatusNotFound,
				Cause:  "CONTEXT_NOT_FOUND",
			}
			return nil, problemDetails
		}
	} else {
		context.UserPool.Range(func(key, value interface{}) bool {
			user := value.(*context.User)
			userInfo := &context.User{
				Id:       user.Id,
				Name:     user.Name,
				Password: user.Password,
			}
			if &userInfo != nil {
				users = append(users, *userInfo)
			}
			return true
		})
		// logger.ServerLog.Infof("Find users: %+v", users)
	}

	return users, nil
}

func HandlePostUser(user *context.User) *httpwrapper.Response {
	context.AddUserToUserPool(user)
	logger.ServerLog.Infof("Add user: %+v", *user)
	return httpwrapper.NewResponse(http.StatusOK, nil, user)
}

func HandleDeleteUser(request *httpwrapper.Request) *httpwrapper.Response {
	id := request.Params["id"]

	if user, ok := context.UserFindById(id); ok {
		logger.ServerLog.Infof("Delete user: %+v", *user)
		context.DeleteUserFromUserPool(id)
		return httpwrapper.NewResponse(http.StatusOK, nil, "Delete user Id="+id)
	} else {
		logger.ServerLog.Warnf("User with Id=%s not found", id)
		problemDetails := &httpwrapper.ProblemDetails{
			Status: http.StatusNotFound,
			Cause:  "CONTEXT_NOT_FOUND",
		}
		return httpwrapper.NewResponse(int(problemDetails.Status), nil, problemDetails)
	}
}

func HandlePutUser(user *context.User) *httpwrapper.Response {

	problemDetails := UpdateUserInformationProcedure(user)
	if problemDetails != nil {
		return httpwrapper.NewResponse(int(problemDetails.Status), nil, problemDetails)
	} else {
		return httpwrapper.NewResponse(http.StatusOK, nil, "Update user context success")
	}
}

func UpdateUserInformationProcedure(user *context.User) *httpwrapper.ProblemDetails {
	id := user.Id

	if curUser, ok := context.UserFindById(id); ok {
		logger.ServerLog.Infof("Update user:\n%+v\nto\n%+v", curUser, user)
		context.AddUserToUserPool(user)
	} else {
		logger.ServerLog.Warnf("Can't update, user with Id=%s not found", id)
		problemDetails := &httpwrapper.ProblemDetails{
			Status: http.StatusNotFound,
			Cause:  "CONTEXT_NOT_FOUND",
		}
		return problemDetails
	}

	return nil
}
