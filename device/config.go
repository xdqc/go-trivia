package device

import (
	"io/ioutil"
	"path/filepath"

	yaml "gopkg.in/yaml.v1"
)

//Config 全局配置
type Config struct {
	Debug          bool   `yaml:"debug"`
	APP            string `yaml:"app"`
	Auto           bool   `yaml:"automatic"`
	Device         string `yaml:"device"`
	OCR            string `yaml:"ocr"`
	AdbAddress     string `yaml:"adb_address"`
	WdaAddress     string `yaml:"wda_address"`
	BaiduAPIKey    string `yaml:"Baidu_API_Key"`
	BaiduSecretKey string `yaml:"Baidu_Secret_Key"`
	BaiduToken     string `yaml:"Baidu_token"`
	BrainID        string `yaml:"Brain_ID"`

	//西瓜视频截图题目位置
	XgQx int `yaml:"xg_q_x"`
	XgQy int `yaml:"xg_q_y"`
	XgQw int `yaml:"xg_q_w"`
	XgQh int `yaml:"xg_q_h"`
	//西瓜视频截取答案位置
	XgAx int `yaml:"xg_a_x"`
	XgAy int `yaml:"xg_a_y"`
	XgAw int `yaml:"xg_a_w"`
	XgAh int `yaml:"xg_a_h"`

	//冲顶大会截图题目位置
	CdQx int `yaml:"cd_q_x"`
	CdQy int `yaml:"cd_q_y"`
	CdQw int `yaml:"cd_q_w"`
	CdQh int `yaml:"cd_q_h"`
	//冲顶大会截取答案位置
	CdAx int `yaml:"cd_a_x"`
	CdAy int `yaml:"cd_a_y"`
	CdAw int `yaml:"cd_a_w"`
	CdAh int `yaml:"cd_a_h"`

	//花椒直播截图题目位置
	HjQx int `yaml:"hj_q_x"`
	HjQy int `yaml:"hj_q_y"`
	HjQw int `yaml:"hj_q_w"`
	HjQh int `yaml:"hj_q_h"`
	//花椒直播截取答案位置
	HjAx int `yaml:"hj_a_x"`
	HjAy int `yaml:"hj_a_y"`
	HjAw int `yaml:"hj_a_w"`
	HjAh int `yaml:"hj_a_h"`

	//芝士超人截图题目位置
	ZsQx int `yaml:"zs_q_x"`
	ZsQy int `yaml:"zs_q_y"`
	ZsQw int `yaml:"zs_q_w"`
	ZsQh int `yaml:"zs_q_h"`
	//芝士超人截取答案位置
	ZsAx int `yaml:"zs_a_x"`
	ZsAy int `yaml:"zs_a_y"`
	ZsAw int `yaml:"zs_a_w"`
	ZsAh int `yaml:"zs_a_h"`

	//nexusq截图题目位置
	NsQx int `yaml:"ns_q_x"`
	NsQy int `yaml:"ns_q_y"`
	NsQw int `yaml:"ns_q_w"`
	NsQh int `yaml:"ns_q_h"`
	//nexusq截取答案位置
	NsAx  int `yaml:"ns_a_x"`
	NsAy  int `yaml:"ns_a_y"`
	NsAw  int `yaml:"ns_a_w"`
	NsAh  int `yaml:"ns_a_h"`
	NsA1x int `yaml:"ns_a1_x"`
	NsA1y int `yaml:"ns_a1_y"`
	NsA1w int `yaml:"ns_a1_w"`
	NsA1h int `yaml:"ns_a1_h"`
	NsA2x int `yaml:"ns_a2_x"`
	NsA2y int `yaml:"ns_a2_y"`
	NsA2w int `yaml:"ns_a2_w"`
	NsA2h int `yaml:"ns_a2_h"`
	NsA3x int `yaml:"ns_a3_x"`
	NsA3y int `yaml:"ns_a3_y"`
	NsA3w int `yaml:"ns_a3_w"`
	NsA3h int `yaml:"ns_a3_h"`
	NsA4x int `yaml:"ns_a4_x"`
	NsA4y int `yaml:"ns_a4_y"`
	NsA4w int `yaml:"ns_a4_w"`
	NsA4h int `yaml:"ns_a4_h"`

	//nexusq_img截图题目位置
	NsiQx int `yaml:"nsi_q_x"`
	NsiQy int `yaml:"nsi_q_y"`
	NsiQw int `yaml:"nsi_q_w"`
	NsiQh int `yaml:"nsi_q_h"`
	//nexusq_img截取答案位置
	NsiAx  int `yaml:"nsi_a_x"`
	NsiAy  int `yaml:"nsi_a_y"`
	NsiAw  int `yaml:"nsi_a_w"`
	NsiAh  int `yaml:"nsi_a_h"`
	NsiA1x int `yaml:"nsi_a1_x"`
	NsiA1y int `yaml:"nsi_a1_y"`
	NsiA1w int `yaml:"nsi_a1_w"`
	NsiA1h int `yaml:"nsi_a1_h"`
	NsiA2x int `yaml:"nsi_a2_x"`
	NsiA2y int `yaml:"nsi_a2_y"`
	NsiA2w int `yaml:"nsi_a2_w"`
	NsiA2h int `yaml:"nsi_a2_h"`
	NsiA3x int `yaml:"nsi_a3_x"`
	NsiA3y int `yaml:"nsi_a3_y"`
	NsiA3w int `yaml:"nsi_a3_w"`
	NsiA3h int `yaml:"nsi_a3_h"`
	NsiA4x int `yaml:"nsi_a4_x"`
	NsiA4y int `yaml:"nsi_a4_y"`
	NsiA4w int `yaml:"nsi_a4_w"`
	NsiA4h int `yaml:"nsi_a4_h"`

	//nexusq_截图sample位置
	NsSPx int `yaml:"ns_sp_x"`
	NsSPy int `yaml:"ns_sp_y"`
	NsSPw int `yaml:"ns_sp_w"`
	NsSPh int `yaml:"ns_sp_h"`
}

var cfg *Config

var cfgFilename = "./config.yml"

//SetConfigFile 设置配置文件地址
func SetConfigFile(path string) {
	cfgFilename = path
}

//GetConfig 解析配置
func GetConfig() *Config {
	if cfg != nil {
		return cfg
	}
	filename, _ := filepath.Abs(cfgFilename)
	yamlFile, err := ioutil.ReadFile(filename)

	if err != nil {
		panic(err)
	}
	var c *Config
	err = yaml.Unmarshal(yamlFile, &c)
	if err != nil {
		panic(err)
	}
	cfg = c
	return cfg
}
