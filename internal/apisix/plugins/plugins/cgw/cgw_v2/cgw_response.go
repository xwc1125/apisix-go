// Package cgw_v2
//
// @author: xwc1125
package cgw_v2

/*
{
    "version":1,
    "componentName":"cgateway",
    "timestamp":1585627862,
    "eventId":"3d45ba5b-a539-4381-8b53-4718b19071f0",
    "requestId":"",
    "returnValue":0,
    "returnCode":0,
    "returnMessage":"",
    "data":{
        "items":[
            {
                "addTimeStamp":"2019-09-04 21:05:37"
            }
        ],
        "totalCount":1
    }
}
*/

// Response defines
type Response struct {
	// httppost.Response
	ComponentName string      `json:"componentName"`
	TimeStamp     int64       `json:"timestamp"`
	RequestId     string      `json:"requestId"`
	ReturnValue   int64       `json:"returnValue"`
	ReturnCode    int64       `json:"returnCode"`
	ReturnMessage string      `json:"returnMessage"`
	Data          interface{} `json:"data"`
}

// NewCgwResponse returns a new Response object
func NewCgwResponse(innerPart interface{}) *Response {
	return &Response{
		Data: innerPart,
	}
}
