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

var EngBalance *engBalance

type engBalance struct {
	EngList     []translator_engine.ITransEngine
	lastIndex   int
	balanceFunc func(i int) translator_engine.ITransEngine
}

func (b *engBalance) GetEng() translator_engine.ITransEngine {
	return EngBalance.balanceFunc(EngBalance.lastIndex)
}

func TransText(eng translator_engine.ITransEngine, src, from, to string) (*translator_engine.TransResult, error) {
	if eng == nil {
		return nil, errors.New("Nonsupport Engine:not initialize")
	}
	return translator_engine.TransText(src, from, to, eng)
}

func BalanceTranText(src, from, to string) (*translator_engine.TransResult, error) {
	if len(EngBalance.EngList) == 0 {
		return nil, errors.New("No translate engine initialized")
	}

	return TransText(EngBalance.GetEng(), src, from, to)

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

	EngBalance = &engBalance{
		EngList:   make([]translator_engine.ITransEngine, 0, 6),
		lastIndex: 0,
		balanceFunc: func(i int) translator_engine.ITransEngine {
			if i < len(EngBalance.EngList)-1 {
				i++
			} else {
				i = 0
			}
			EngBalance.lastIndex = i
			return EngBalance.EngList[i]
		},
	}

	youdaoAppKey, youdaoSecretKey := getYoudaoCfg()
	if YoudaoEng = translator_engine.EngFactory.BuildYoudaoEng(youdaoAppKey, youdaoSecretKey); YoudaoEng != nil {
		EngBalance.EngList = append(EngBalance.EngList, YoudaoEng)
	}

	tcSecretId, tcSecretKey := getTencentCfg()
	if TencentEng, _ = translator_engine.EngFactory.BuildTencentEng(tcSecretId, tcSecretKey); TencentEng != nil {
		EngBalance.EngList = append(EngBalance.EngList, TencentEng)
	}

	baiduApiKey, baiduSecretKey := getBaiduCfg()
	if BaiduEng = translator_engine.EngFactory.BuildBaiduEng(baiduApiKey, baiduSecretKey); BaiduEng != nil {
		EngBalance.EngList = append(EngBalance.EngList, BaiduEng)
	}

	aliAccessId, aliAccessSecret := getAliCfg()
	if AliEng, _ = translator_engine.EngFactory.BuildAliEngine(aliAccessId, aliAccessSecret); AliEng != nil {
		EngBalance.EngList = append(EngBalance.EngList, AliEng)
	}

	volcAccessKey, volcAccessScret := getVolcConfig()
	if VolcEng = translator_engine.EngFactory.BuildVolcEngine(volcAccessKey, volcAccessScret); VolcEng != nil {
		EngBalance.EngList = append(EngBalance.EngList, VolcEng)
	}

}
