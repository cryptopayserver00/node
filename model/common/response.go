package common

import "github.com/gin-gonic/gin"

type CODE_INT int

type Response struct {
	Code    CODE_INT `json:"code"`
	Message string   `json:"message"`
	Data    any      `json:"data"`
	Result  bool     `json:"result"`
}

const (
	Success               CODE_INT = 10200
	Success_OK            CODE_INT = 10200
	Success_Created       CODE_INT = 10201
	Success_Accepted      CODE_INT = 10202
	Success_No_Content    CODE_INT = 10204
	Success_Reset_Content CODE_INT = 10205

	Success_Msg = "Request data successful"
)

const (
	Error                        CODE_INT = 10500
	Error_Bad_Request            CODE_INT = 10400
	Error_Unauthorized           CODE_INT = 10401
	Error_Forbidden              CODE_INT = 10403
	Error_Not_Found              CODE_INT = 10404
	Error_Not_Acceptable         CODE_INT = 10406
	Error_Resource_Conflict      CODE_INT = 10409
	Error_Unsupported_Media_Type CODE_INT = 10415
	Error_Unprocessable_Entity   CODE_INT = 10422
	Error_Locked                 CODE_INT = 10423
	Error_Too_Many_Requests      CODE_INT = 10429
	Error_Internal_Server_Error  CODE_INT = 10500
	Error_Not_Implemented        CODE_INT = 10501
	Error_Bad_Gateway            CODE_INT = 10502
	Error_Service_Unavailable    CODE_INT = 10503
	Error_Gateway_Timeout        CODE_INT = 10504
	Error_Origin_DNS_Error       CODE_INT = 10530
	Error_Unexpected_Token       CODE_INT = 10783

	Failed_Msg = "Request data failed"
)

func Result(code CODE_INT, msg string, data any, result bool) Response {
	return Response{
		code,
		msg,
		data,
		result,
	}
}

func Ok(c *gin.Context) Response {
	return Result(Success, Success_Msg, map[string]any{}, true)
}

func OkWithMessage(message string) Response {
	return Result(Success, message, map[string]any{}, true)
}

func OKWithData(data any) Response {
	return Result(Success, Success_Msg, data, true)
}

func OkWithDetailed(code CODE_INT, message string, data any) Response {
	return Result(code, message, data, true)
}

func Fail() Response {
	return Result(Error, Failed_Msg, map[string]any{}, false)
}

func FailWithMessage(message string) Response {
	return Result(Error, message, map[string]any{}, false)
}

func FailWithDetailed(code CODE_INT, message string, data any) Response {
	return Result(code, message, data, false)
}
