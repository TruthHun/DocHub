package AdminControllers

import "github.com/TruthHun/DocHub/models"

type ScoreController struct {
	BaseController
}

//积分管理
func (this *ScoreController) Get() {
	var log models.CoinLog
	log.Uid, _ = this.GetInt("uid")
	log.Coin, _ = this.GetInt("score")
	log.Log = this.GetString("log")
	err := models.Regulate(models.TableUserInfo, "Coin", log.Coin, "Id=?", log.Uid)
	if err == nil {
		err = models.ModelCoinLog.LogRecord(log)
	}
	if err != nil {
		this.ResponseJson(0, err.Error())
	} else {
		this.ResponseJson(1, "积分变更成功")
	}
}
