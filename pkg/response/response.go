package response

import "github.com/gin-gonic/gin"

type Response struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data,omitempty"`
}

func JSON(c *gin.Context, status int, message string, data any) {
	c.JSON(status, Response{Code: status, Msg: message, Data: data})
}
