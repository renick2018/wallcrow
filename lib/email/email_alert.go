package email

import (
	"fmt"
	"wallcrow/lib"
	"wallcrow/lib/logger"
)

func Alert(title, content string)  {
	for _, item := range lib.Global.Emails {
		logger.Info(fmt.Sprintf("send email alert %s to %s res: %v", title, item, Send(item, title, content)))
	}
}
