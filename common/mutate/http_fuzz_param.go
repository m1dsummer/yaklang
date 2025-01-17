package mutate

import (
	"fmt"
	"github.com/asaskevich/govalidator"
	"github.com/yaklang/yaklang/common/go-funk"
	"github.com/yaklang/yaklang/common/jsonextractor"
	"github.com/yaklang/yaklang/common/log"
	"github.com/yaklang/yaklang/common/utils"
	"strings"
)

type httpParamPositionType string

var (
	posMethod              httpParamPositionType = "method"
	posBody                httpParamPositionType = "body"
	posGetQuery            httpParamPositionType = "get-query"
	posGetQueryJson        httpParamPositionType = "get-query-json"
	posGetQueryBase64Json  httpParamPositionType = "get-query-base64-json"
	posPath                httpParamPositionType = "path"
	posHeader              httpParamPositionType = "header"
	posPostQuery           httpParamPositionType = "post-query"
	posPostQueryJson       httpParamPositionType = "post-query-json"
	posPostQueryBase64Json httpParamPositionType = "post-query-base64-json"
	posPostJson            httpParamPositionType = "post-json"
	posCookie              httpParamPositionType = "cookie"
	posCookieJson          httpParamPositionType = "cookie-json"
	posCookieBase64Json    httpParamPositionType = "cookie-base64-json"
	posPathAppend          httpParamPositionType = "path-append"
	posPathBlock           httpParamPositionType = "path-block"
)

func PositionTypeVerbose(pos httpParamPositionType) string {
	switch pos {
	case posMethod:
		return "HTTP方法"
	case posBody:
		return "Body"
	case posGetQuery:
		return "GET参数"
	case posGetQueryJson:
		return "GET参数(JSON)"
	case posGetQueryBase64Json:
		return "GET参数(Base64+JSON)"
	case posPathAppend:
		return "URL路径(追加)"
	case posPathBlock:
		return "URL路径(分块)"
	case posPath:
		return "URL路径"
	case posHeader:
		return "Header"
	case posPostQuery:
		return "POST参数"
	case posPostQueryJson:
		return "POST参数(JSON)"
	case posPostQueryBase64Json:
		return "POST参数(Base64+JSON)"
	case posPostJson:
		return "JSON-Body参数"
	case posCookie:
		return "Cookie参数"
	case posCookieJson:
		return "Cookie参数(JSON)"
	case posCookieBase64Json:
		return "Cookie参数(Base64+JSON)"
	default:
		return string(pos)
	}
}

type FuzzHTTPRequestParam struct {
	typePosition     httpParamPositionType
	param            interface{}
	param2nd         interface{}
	paramOriginValue interface{}
	jsonPath         string
	origin           *FuzzHTTPRequest
}

func (p *FuzzHTTPRequestParam) IsPostParams() bool {
	if p.typePosition == posPostJson {
		return true
	}

	if p.typePosition == posPostQuery {
		return true
	}

	return false
}

func (p *FuzzHTTPRequestParam) IsGetParams() bool {
	if p.typePosition == posGetQuery {
		return true
	}

	return false
}

func (p *FuzzHTTPRequestParam) IsGetValueJSON() bool {
	if p == nil {
		return false
	}

	if !p.IsGetParams() {
		return false
	}

	valStr := utils.InterfaceToString(utils.InterfaceToString(p.Value()))
	fixedJson := jsonextractor.FixJson([]byte(valStr))
	return govalidator.IsJSON(string(fixedJson))
}

func (p *FuzzHTTPRequestParam) IsCookieParams() bool {
	if p.typePosition == posCookie {
		return true
	}

	return false
}

func (p *FuzzHTTPRequestParam) Name() string {
	if p.param2nd != nil {
		return ""
	}
	return fmt.Sprintf("%v", p.param)
}

func (p *FuzzHTTPRequestParam) Position() string {
	return string(p.typePosition)
}

func (p *FuzzHTTPRequestParam) PositionVerbose() string {
	return PositionTypeVerbose(p.typePosition)
}

func (p *FuzzHTTPRequestParam) Value() interface{} {
	return p.paramOriginValue
}

func (p *FuzzHTTPRequestParam) Repeat(i int) FuzzHTTPRequestIf {
	return p.origin.Repeat(i)
}

func (p *FuzzHTTPRequestParam) Fuzz(i ...interface{}) FuzzHTTPRequestIf {
	switch p.typePosition {
	case posMethod:
		return p.origin.FuzzMethod(InterfaceToFuzzResults(i)...)
	case posGetQuery:
		return p.origin.FuzzGetParams(p.param, i)
	case posGetQueryJson:
		return p.origin.FuzzGetJsonPathParams(p.param, p.jsonPath, i)
	case posGetQueryBase64Json:
		return p.origin.FuzzGetBase64JsonPath(p.param, p.jsonPath, i)
	case posHeader:
		return p.origin.FuzzHTTPHeader(p.param, i)
	case posPath:
		return p.origin.FuzzPath(InterfaceToFuzzResults(i)...)
	case posPostJson:
		return p.origin.FuzzPostJsonParams(p, i)
	case posCookie:
		return p.origin.FuzzCookie(p.param, InterfaceToFuzzResults(i))
	case posCookieJson:
		return p.origin.FuzzCookieJsonPath(p.param, p.jsonPath, i)
	case posCookieBase64Json:
		return p.origin.FuzzCookieBase64JsonPath(p.param, p.jsonPath, i)
	case posPostQuery:
		return p.origin.FuzzPostParams(p.param, i)
	case posPostQueryJson:
		return p.origin.FuzzPostJsonPathParams(p.param, p.jsonPath, i)
	case posPostQueryBase64Json:
		return p.origin.FuzzPostBase64JsonPath(p.param, p.jsonPath, i)
	case posPathAppend:
		return p.origin.FuzzPath(funk.Map(InterfaceToFuzzResults(i), func(s string) string {
			if !strings.HasPrefix(s, "/") {
				s = "/" + s
			}
			return p.origin.GetPath() + s
		}).([]string)...)
	case posBody:
		return p.origin.FuzzPostRaw(InterfaceToFuzzResults(i)...)
	case posPathBlock:
		var result = strings.Split(p.origin.GetPath(), "/")
		if len(result) <= 0 {
			return p.origin.FuzzPath(InterfaceToFuzzResults(i)...)
		}
		var templates []string
		for i := 1; i < len(result); i++ {
			resultCopy := result[:]
			resultCopy[i] = `{{params(placeholder)}}`
			templates = append(templates, strings.Join(resultCopy, "/"))
		}
		fuzzResults := InterfaceToFuzzResults(i)
		var finalResults []string
		for _, t := range templates {
			finalResults = append(finalResults, InterfaceToFuzzResults(t, MutateWithExtraParams(map[string][]string{
				"placeholder": fuzzResults,
			}))...)
		}
		return p.origin.FuzzPath(finalResults...)
	default:
		log.Warnf("cannot found fuzz params method identify: %v", posGetQueryJson)
		return p.origin
	}
}

func (p *FuzzHTTPRequestParam) String() string {
	if p.jsonPath != "" {
		return fmt.Sprintf("Name:%-20s JsonPath: %-12s Position:[%v(%v)]\n", p.Name(), p.jsonPath, p.PositionVerbose(), p.Position())
	}
	return fmt.Sprintf("Name:%-20s Position:[%v(%v)]\n", p.Name(), p.PositionVerbose(), p.Position())
}

func (p *FuzzHTTPRequestParam) Debug() *FuzzHTTPRequestParam {
	fmt.Print(p.String())
	return p
}
