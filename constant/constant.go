package constant

import "fmt"

const BaseUrl = "https://www.aeropres.in/chromeapi/dawn/v1"

var KeepAliveURL = fmt.Sprintf("%v/userreward/keepalive", BaseUrl)
var GetPointURL = "https://www.aeropres.in/api/atom/v1/userreferral/getpoint"
var LoginURL = "https://www.aeropres.in/chromeapi/dawn/v1/user/login"
