package handler

import (
	"net/http"

	"github.com/devnull-twitch/pond-com/protobuf/com/v1"
	"github.com/devnull-twitch/pond/pkg/auth"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func (h *handler) AuthHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Query("code") == "" {
			logrus.Warn("missing code param")
			if err := h.render(c, http.StatusBadRequest, "tpl/auth/error.html.tmpl", nil); err != nil {
				panic(err)
			}
			return
		}
		if c.Query("state") == "" {
			logrus.Warn("missing state param")
			if err := h.render(c, http.StatusBadRequest, "tpl/auth/error.html.tmpl", nil); err != nil {
				panic(err)
			}
			return
		}
		reqToken := c.Query("state")

		accessResp, err := h.twClient.RequestUserAccessToken(c.Query("code"))
		if err != nil {
			logrus.WithError(err).Error("unable to generate user access token")
			if err := h.render(c, http.StatusBadRequest, "tpl/auth/error.html.tmpl", nil); err != nil {
				panic(err)
			}
			return
		}
		if err := h.validateTwitchResponseCommon(c, &accessResp.ResponseCommon); err != nil {
			if err := h.render(c, http.StatusBadRequest, "tpl/auth/error.html.tmpl", nil); err != nil {
				panic(err)
			}
			return
		}

		tokenOk, validateResponse, err := h.twClient.ValidateToken(accessResp.Data.AccessToken)
		if !tokenOk {
			logrus.Error("invalid twitch access token")
			if err := h.render(c, http.StatusBadRequest, "tpl/auth/error.html.tmpl", nil); err != nil {
				panic(err)
			}
			return
		}
		if err != nil {
			logrus.WithError(err).Error("unable to validate user access token")
			if err := h.render(c, http.StatusBadRequest, "tpl/auth/error.html.tmpl", nil); err != nil {
				panic(err)
			}
			return
		}
		if err := h.validateTwitchResponseCommon(c, &validateResponse.ResponseCommon); err != nil {
			if err := h.render(c, http.StatusBadRequest, "tpl/auth/error.html.tmpl", nil); err != nil {
				panic(err)
			}
			return
		}

		jwtString, err := auth.CreateToken(validateResponse.Data.Login, accessResp.Data.AccessToken, accessResp.Data.RefreshToken)
		if err != nil {
			logrus.WithError(err).Error("unable to validate user access token")
			if err := h.render(c, http.StatusBadRequest, "tpl/auth/error.html.tmpl", nil); err != nil {
				panic(err)
			}
			return
		}

		auth.Set(reqToken, com.PollResponse_STATUS_SUCCESS, &jwtString)
		h.render(c, http.StatusOK, "tpl/auth/success.html.tmpl", nil)
	}
}
