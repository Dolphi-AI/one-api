package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/model"
)

func authHelper(c *gin.Context, minRole int) {
	session := sessions.Default(c)
	username := session.Get("username")
	role := session.Get("role")
	id := session.Get("id")
	status := session.Get("status")
	if username == nil {
		logger.SysLog("no user session found, try to use access token")
		// Check access token
		accessToken := c.Request.Header.Get("Authorization")
		if accessToken == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "无权进行此操作，未登录且未提供 access token",
			})
			c.Abort()
			return
		}

		user := model.ValidateAccessToken(accessToken)
		if user != nil && user.Username != "" {
			// Token is valid
			username = user.Username
			role = user.Role
			id = user.Id
			status = user.Status
		} else {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "无权进行此操作，access token 无效",
			})
			c.Abort()
			return
		}
	}

	if status.(int) == common.UserStatusDisabled {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "用户已被封禁",
		})
		c.Abort()
		return
	}
	if role.(int) < minRole {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无权进行此操作，权限不足",
		})
		c.Abort()
		return
	}
	c.Set("username", username)
	c.Set("role", role)
	c.Set("id", id)
	c.Next()
}

func UserAuth() func(c *gin.Context) {
	return func(c *gin.Context) {
		authHelper(c, common.RoleCommonUser)
	}
}

func AdminAuth() func(c *gin.Context) {
	return func(c *gin.Context) {
		authHelper(c, common.RoleAdminUser)
	}
}

func RootAuth() func(c *gin.Context) {
	return func(c *gin.Context) {
		authHelper(c, common.RoleRootUser)
	}
}

func TokenAuth() func(c *gin.Context) {
	return func(c *gin.Context) {
		key := c.Request.Header.Get("Authorization")
		key = strings.TrimPrefix(key, "Bearer ")
		key = strings.TrimPrefix(strings.TrimPrefix(key, "sk-"), "laisky-")
		parts := strings.Split(key, "-")
		key = parts[0]
		token, err := model.ValidateUserToken(key)
		if err != nil {
			abortWithMessage(c, http.StatusUnauthorized, err.Error())
			return
		}
		userEnabled, err := model.CacheIsUserEnabled(token.UserId)
		if err != nil {
			abortWithMessage(c, http.StatusInternalServerError, err.Error())
			return
		}
		if !userEnabled {
			abortWithMessage(c, http.StatusForbidden, "用户已被封禁")
			return
		}
		c.Set("id", token.UserId)
		c.Set("token_id", token.Id)
		c.Set("token_name", token.Name)
		if len(parts) > 1 {
			if model.IsAdmin(token.UserId) {
				c.Set("specific_channel_id", parts[1])
			} else {
				abortWithMessage(c, http.StatusForbidden, "普通用户不支持指定渠道")
				return
			}
		}
		c.Next()
	}
}
