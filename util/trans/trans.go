package trans

import (
	translator_engine "github.com/NoahAmethyst/translator-engine"
	"github.com/pkg/errors"
	"os"
)

var YoudaoEng *translator_engine.YoudaoTransEngine
var TencentEng *translator_engine.TencentTransEngine
var BaiduEng *translator_engine.BaiduTransEngine
var AliEng *translator_engine.AliTransEngine
var VolcEng *translator_engine.VolcTransEngine
var EngList []translator_engine.ITransEngine

var currIndex int

func TransText(eng translator_engine.ITransEngine, src, from, to string) (*translator_engine.TransResult, error) {
	if eng == nil {
		return nil, errors.New("Nonsupport Engine:not initialize")
	}
	return translator_engine.TransText(src, from, to, eng)
}

func BalanceTranText(src, from, to string) (*translator_engine.TransResult, error) {
	if len(EngList) == 0 {
		return nil, errors.New("No translate engine initialized")
	}

	defer func() {
		currIndex++
	}()
	i := currIndex % len(EngList)
	return TransText(EngList[i], src, from, to)

}

func getBaiduCfg() (string, string) {
	return os.Getenv("BAIDU_API_KEY"), os.Getenv("BAIDU_SECRET_KEY")
}

func getAliCfg() (string, string) {
	return os.Getenv("ALI_ACCESS_ID"), os.Getenv("ALI_ACCESS_SECRET")
}

func getTencentCfg() (string, string) {
	return os.Getenv("TC_SECRET_ID"), os.Getenv("TC_SECRET_KEY")
}
func getYoudaoCfg() (string, string) {
	return os.Getenv("YD_APP_KEY"), os.Getenv("YD_SECRET_KEY")
}

func getVolcConfig() (string, string) {
	return os.Getenv("VOLC_ACCESS_KEY"), os.Getenv("VOLC_SECRET_KEY")
}

func init() {

	youdaoAppKey, youdaoSecretKey := getYoudaoCfg()
	if YoudaoEng = translator_engine.EngFactory.BuildYoudaoEng(youdaoAppKey, youdaoSecretKey); YoudaoEng != nil {
		EngList = append(EngList, YoudaoEng)
	}

	tcSecretId, tcSecretKey := getTencentCfg()
	if TencentEng, _ = translator_engine.EngFactory.BuildTencentEng(tcSecretId, tcSecretKey); TencentEng != nil {
		EngList = append(EngList, TencentEng)
	}

	baiduApiKey, baiduSecretKey := getBaiduCfg()
	if BaiduEng = translator_engine.EngFactory.BuildBaiduEng(baiduApiKey, baiduSecretKey); BaiduEng != nil {
		EngList = append(EngList, BaiduEng)
	}

	aliAccessId, aliAccessSecret := getAliCfg()
	if AliEng, _ = translator_engine.EngFactory.BuildAliEngine(aliAccessId, aliAccessSecret); AliEng != nil {
		EngList = append(EngList, AliEng)
	}

	volcAccessKey, volcAccessScret := getVolcConfig()
	if VolcEng = translator_engine.EngFactory.BuildVolcEngine(volcAccessKey, volcAccessScret); VolcEng != nil {
		EngList = append(EngList, VolcEng)
	}

}
