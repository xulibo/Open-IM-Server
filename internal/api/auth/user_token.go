package apiAuth

import (
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/grpc-etcdv3/getcdv3"
	pbAuth "Open_IM/pkg/proto/auth"
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

// paramsUserToken struct
type paramsUserToken struct {
	Secret   string `json:"secret" binding:"required,max=32"`
	Platform int32  `json:"platform" binding:"required,min=1,max=8"`
	UID      string `json:"uid" binding:"required,min=1,max=64"`
}

func newUserTokenReq(params *paramsUserToken) *pbAuth.UserTokenReq {
	pbData := pbAuth.UserTokenReq{
		Platform: params.Platform,
		UID:      params.UID,
	}
	return &pbData
}

// @Summary
// @Schemes
// @Description get token
// @Tags auth
// @Accept json
// @Produce json
// @Param body body apiAuth.paramsUserToken true "get token params"
// @Success 200 {object} user.result{data=apiAuth.resultUserRegister}
// @Failure 400 {object} user.result
// @Failure 500 {object} user.result
// @Router /auth/user_token [post]
func UserToken(c *gin.Context) {
	log.Info("", "", "api user_token init ....")
	etcdConn := getcdv3.GetAuthConn()
	client := pbAuth.NewAuthClient(etcdConn)
	//defer etcdConn.Close()

	params := paramsUserToken{}
	if err := c.BindJSON(&params); err != nil {
		log.Error("", "", params.UID, params.Platform, params.Secret)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}
	if params.Secret != config.Config.Secret {
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 401, "errMsg": "not authorized"})
		return
	}
	pbData := newUserTokenReq(&params)

	log.Info("", "", "api user_token is server, [data: %s]", pbData.String())
	reply, err := client.UserToken(context.Background(), pbData)
	if err != nil {
		log.Error("", "", "api user_token call rpc fail, [data: %s] [err: %s]", pbData.String(), err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"errCode": 500, "errMsg": err.Error()})
		return
	}
	log.Info("", "", "api user_token call rpc success, [data: %s] [reply: %s]", pbData.String(), reply.String())

	if reply.ErrCode == 0 {
		c.JSON(http.StatusOK, gin.H{
			"errCode": reply.ErrCode,
			"errMsg":  reply.ErrMsg,
			"data": gin.H{
				"uid":         pbData.UID,
				"token":       reply.Token,
				"expiredTime": reply.ExpiredTime,
			},
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"errCode": reply.ErrCode,
			"errMsg":  reply.ErrMsg,
		})
	}

}