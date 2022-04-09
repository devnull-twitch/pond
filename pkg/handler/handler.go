package handler

import (
	"bytes"
	"fmt"
	"html/template"

	"github.com/gin-gonic/gin"
	"github.com/nicklaw5/helix"
	"github.com/sirupsen/logrus"
)

type Handler interface {
	AuthHandler() gin.HandlerFunc
}

type handler struct {
	twClient  *helix.Client
	baseTmpl  *template.Template
	tmplCache map[string]*template.Template
}

type twCommonError struct {
	Status  int
	Message string
}

func (tce twCommonError) Error() string {
	return fmt.Sprintf("twitch api error [%d] %s", tce.Status, tce.Message)
}

func New(twClient *helix.Client) Handler {
	baseTmpl := template.New("AuthHandler")
	baseTmpl.ParseGlob("tpl/layout/*.html.tmpl")

	return &handler{
		twClient:  twClient,
		baseTmpl:  baseTmpl,
		tmplCache: make(map[string]*template.Template),
	}
}

func (h *handler) validateTwitchResponseCommon(c *gin.Context, responseCommon *helix.ResponseCommon) error {
	if responseCommon.ErrorStatus > 0 {
		logrus.WithFields(logrus.Fields{
			"tw_status":  responseCommon.StatusCode,
			"tw_message": responseCommon.ErrorMessage,
		}).Error("twitch api error")
		return twCommonError{Status: responseCommon.StatusCode, Message: responseCommon.ErrorMessage}
	}

	return nil
}

func (h *handler) render(c *gin.Context, status int, template string, templateInput interface{}) error {
	if _, ok := h.tmplCache[template]; !ok {
		clone, err := h.baseTmpl.Clone()
		if err != nil {
			return err
		}
		clone.ParseFiles(template)
		logrus.WithField("template", template).Info("parsed template")
		h.tmplCache[template] = clone
	}

	bufWriter := bytes.NewBuffer([]byte{})
	if err := h.tmplCache[template].Execute(bufWriter, templateInput); err != nil {
		return err
	}

	c.Data(status, gin.MIMEHTML, bufWriter.Bytes())
	return nil
}
