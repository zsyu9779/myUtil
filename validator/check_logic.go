package validator

import (
	"github.com/dlclark/regexp2"
	"regexp"
)

//业务逻辑校验

// 手机号校验
func CheckMobile(phne string) bool {
	//匹配规则
	regRuler := "^1[3456789]{1}\\d{9}$"
	//正则调用
	reg := regexp.MustCompile(regRuler)
	return reg.MatchString(phne)
}

// 身份证号验证
func CheckIdCard(card string) bool {
	regRuler := "(^\\d{15}$)|(^\\d{18}$)|(^\\d{17}(\\d|X|x)$)"
	reg := regexp.MustCompile(regRuler)
	return reg.MatchString(card)
}

// 验证邮箱
func CheckEmail(email string) bool {
	regRuler := "^[A-Z0-9._%+-]+@[A-Z0-9.-]+\\.[A-Z]{2,6}"
	reg := regexp.MustCompile(regRuler)
	return reg.MatchString(email)
}

// 验证密码输入是否合规
func Check3Password(str string) bool {
	//密码必须为8-16位数字/字母/特殊字符的组合
	expr := `^(?![0-9a-zA-Z]+$)(?![a-zA-Z!@#$%^&*]+$)(?![0-9!@#$%^&*]+$)[0-9A-Za-z!@#$%^&*]{8,16}$`
	reg, _ := regexp2.Compile(expr, 0)
	m, _ := reg.FindStringMatch(str)
	if m != nil {
		//res:= m.String()
		return true
	}
	return false
}

// 验证密码输入是否合规
func Check2Password(str string) bool {
	//密码必须为8-16位数字和字母的组合
	expr := `^(?![a-zA-Z!@#$%^&*]+$)(?![0-9!@#$%^&*]+$)[0-9A-Za-z!@#$%^&*]{8,16}$`
	reg, _ := regexp2.Compile(expr, 0)
	m, _ := reg.FindStringMatch(str)
	if m != nil {
		//res:= m.String()
		return true
	}
	return false
}

// 校验用户名称
func CheckNickname(str string) bool {
	expr := `[/'!*$]+`
	reg, _ := regexp2.Compile(expr, 0)
	m, _ := reg.FindStringMatch(str)
	if m != nil {
		return true
	}
	return false
}
