package ecode

// define your code here
const (
	// common error code
	OK              ecode = 0
	AppKeyInvalid   ecode = -1   // 应用程序不存在或已被封禁
	AccessKeyErr    ecode = -2   // Access Key错误
	SignCheckErr    ecode = -3   // API校验密匙错误
	NoLogin         ecode = -101 // 账号未登录
	UserDisabled    ecode = -102 // 账号被封停
	LackOfScores    ecode = -103 // 积分不足
	LackOfCoins     ecode = -104 // 硬币不足
	CaptchaErr      ecode = -105 // 验证码错误
	UserInactive    ecode = -106 // 账号未激活
	UserNoMember    ecode = -107 // 账号非正式会员或在适应期
	AppDenied       ecode = -108 // 应用不存在或者被封禁
	MobileNoVerfiy  ecode = -110 // 未绑定手机
	CsrfNotMatchErr ecode = -111 // csrf 校验失败
	ServiceUpdate   ecode = -112 // 系统升级中

	RequestErr ecode = -400 //
	ServerErr  ecode = -500 // 服务器错误

	// reply
	ReplyNotExist ecode = 10000

	// user
	UserNotExist ecode = 11000
)

var (
	ecodeMap = map[ecode]string{
		OK:         "ok",
		ServerErr:  "服务器错误",
		RequestErr: "参数错误",
		// 评论
		ReplyNotExist: "评论不存在",
		// 用户
		UserNotExist: "用户不存在",
	}
)
