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
	err := models.Regulate(models.GetTableUserInfo(), "Coin", log.Coin, "Id=?", log.Uid)
	if err == nil {
		err = models.NewCoinLog().LogRecord(log)
	}
	if err != nil {
		this.ResponseJson(false, err.Error())
	}
	this.ResponseJson(true, "积分变更成功")
}
