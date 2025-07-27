// Package smtp
package smtp

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net/smtp"
	"strings"
	"time"

	"github.com/iconnor-code/cogo/core"

	"go.uber.org/zap"
)

type EmailSMTP struct {
	host     string
	port     int
	username string
	password string
	logger   core.ILogger

	from    string
	to      []string
	msg     []byte
	subject string
}

func NewSMTP(conf core.IConfig, logger core.ILogger) *EmailSMTP {
	return &EmailSMTP{
		host:     conf.Get("smtp.host").(string),
		port:     conf.Get("smtp.port").(int),
		username: conf.Get("smtp.username").(string),
		password: conf.Get("smtp.password").(string),
		logger:   logger,
	}
}

func (e *EmailSMTP) SendVerifyCode(ctx context.Context, from string, to string, code string, period time.Duration) error {
	e.from = from
	e.to = []string{to}
	e.subject = "您的验证码"

	body := fmt.Sprintf("您的验证码是: %s 有效期: %.0f分钟", code, period.Minutes())
	encodedAppName := "=?UTF-8?B?" + base64.StdEncoding.EncodeToString([]byte(e.from)) + "?="
	encodedSubject := "=?UTF-8?B?" + base64.StdEncoding.EncodeToString([]byte(e.subject)) + "?="

	e.msg = []byte("From: " + encodedAppName + " <" + e.username + ">\r\n" +
		"To: " + strings.Join(e.to, ",") + "\r\n" +
		"Subject: " + encodedSubject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/plain; charset=UTF-8\r\n" +
		"Content-Transfer-Encoding: 8bit\r\n" +
		"\r\n" +
		body)

	go e.sendEmail()

	return nil
}

func (e *EmailSMTP) sendEmail() {
	addr := fmt.Sprintf("%s:%d", e.host, e.port)
	tlsConfig := &tls.Config{
		ServerName:         e.host,
		InsecureSkipVerify: false,
	}

	conn, err := tls.Dial("tcp", addr, tlsConfig)
	e.log()

	if err != nil {
		e.logger.Error("SMTP建立TLS连接失败", zap.Error(err))
		return
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, e.host)
	if err != nil {
		e.logger.Error("SMTP创建客户端失败", zap.Error(err))
		return
	}
	defer client.Close()

	auth := smtp.PlainAuth("", e.username, e.password, e.host)
	if err = client.Auth(auth); err != nil {
		e.logger.Error("SMTP认证失败", zap.Error(err))
		return
	}

	if err = client.Mail(e.username); err != nil {
		e.logger.Error("SMTP设置发件人失败", zap.Error(err))
		return
	}

	for _, rcptAddr := range e.to {
		if err = client.Rcpt(rcptAddr); err != nil {
			e.logger.Error("SMTP设置收件人失败", zap.Error(err))
			return
		}
	}

	w, err := client.Data()
	if err != nil {
		e.logger.Error("SMTP创建邮件内容写入器失败", zap.Error(err))
		return
	}

	_, err = w.Write(e.msg)
	if err != nil {
		e.logger.Error("SMTP写入邮件内容失败", zap.Error(err))
		return
	}

	err = w.Close()
	if err != nil {
		e.logger.Error("SMTP关闭邮件内容写入器失败", zap.Error(err))
		return
	}

	err = client.Quit()
	if err != nil {
		e.logger.Error("SMTP关闭连接失败", zap.Error(err))
		return
	}

	e.logger.Info("SMTP发送邮件完成")
}

func (e *EmailSMTP) log() {
	e.logger.AddGlobalFields(
		zap.String("app_name", e.from),
		zap.String("host", e.host),
		zap.Int("port", e.port),
		zap.String("from", e.username),
		zap.String("to", strings.Join(e.to, ",")),
		zap.String("subject", e.subject),
		zap.String("msg", string(e.msg)),
	)
}
