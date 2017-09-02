package service

import "errors"

var (
	PwdWrongError = errors.New("密码错误")
	NotExistError = errors.New("找不到对象")
	TimeOutError  = errors.New("超时")
)
