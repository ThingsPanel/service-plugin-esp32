// internal/handler/handler.go
package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	formjson "tp-plugin/internal/form_json"
	"tp-plugin/internal/platform"

	"github.com/ThingsPanel/tp-protocol-sdk-go/handler"
	"github.com/sirupsen/logrus"
)

// logrusWriter 实现 io.Writer 接口用于适配logrus
type logrusWriter struct {
	logger *logrus.Logger
}

func (w *logrusWriter) Write(p []byte) (n int, err error) {
	w.logger.Info(string(p))
	return len(p), nil
}

// HTTPHandler HTTP服务处理器
type HTTPHandler struct {
	platform *platform.PlatformClient
	logger   *logrus.Logger
	stdlog   *log.Logger
}

// NewHTTPHandler 创建HTTP处理器
func NewHTTPHandler(platform *platform.PlatformClient, logger *logrus.Logger) *HTTPHandler {
	// 创建适配器
	writer := &logrusWriter{logger: logger}
	stdlog := log.New(writer, "[HTTP] ", log.Ldate|log.Ltime|log.Lshortfile)

	return &HTTPHandler{
		platform: platform,
		logger:   logger,
		stdlog:   stdlog,
	}
}

// RegisterHandlers 注册所有HTTP处理器
func (h *HTTPHandler) RegisterHandlers() *handler.Handler {
	// 创建处理器，使用标准库Logger
	hdl := handler.NewHandler(handler.HandlerConfig{
		Logger: h.stdlog,
	})

	// 设置表单配置处理函数
	hdl.SetFormConfigHandler(h.handleGetFormConfig)

	// 设置设备断开连接处理函数
	hdl.SetDeviceDisconnectHandler(h.handleDeviceDisconnect)

	// 设置通知处理函数
	hdl.SetNotificationHandler(h.handleNotification)

	// 设置获取设备列表处理函数
	hdl.SetGetDeviceListHandler(h.handleGetDeviceList)

	// 设置获取设备详细处理函数
	hdl.SetGetDeviceInfoHandler(h.handleGetDeviceInfo)

	return hdl
}

// handleGetFormConfig 处理获取表单配置请求
func (h *HTTPHandler) handleGetFormConfig(req *handler.GetFormConfigRequest) (interface{}, error) {
	h.logger.WithFields(logrus.Fields{
		"protocol_type": req.ProtocolType,
		"device_type":   req.DeviceType,
		"form_type":     req.FormType,
	}).Info("收到获取表单配置请求")

	// 根据请求类型返回不同的配置表单
	switch req.FormType {
	case "CFG": // 设备配置表单
		return nil, nil
	case "VCR": // 设备凭证表单
		return nil, nil
	case "SVCR": // 服务接入点凭证表单
		return readFormConfigByPath("../internal/form_json/form_service_voucher.json"), nil
	default:
		return nil, fmt.Errorf("不支持的表单类型: %s", req.FormType)
	}
}

// ./form_config.json
func readFormConfigByPath(path string) interface{} {
	filePtr, err := os.Open(path)
	if err != nil {
		logrus.Warn("文件打开失败...", err.Error())
		return nil
	}
	defer filePtr.Close()
	var info interface{}
	// 创建json解码器
	decoder := json.NewDecoder(filePtr)
	err = decoder.Decode(&info)
	if err != nil {
		logrus.Warn("解码失败", err.Error())
		return info
	} else {
		logrus.Info("读取文件[form_config.json]成功...")
		return info
	}
}

// handleDeviceDisconnect 处理设备断开连接请求
func (h *HTTPHandler) handleDeviceDisconnect(req *handler.DeviceDisconnectRequest) error {
	h.logger.WithField("device_id", req.DeviceID).Info("收到设备断开连接请求")

	// 清理设备缓存
	// Note: 因为原缓存是按 device_number 存储的,这里要先查出设备信息
	device, err := h.platform.GetDeviceByID(req.DeviceID)
	if err == nil { // 如果能找到设备就清理缓存
		h.platform.ClearDeviceCache(device.DeviceNumber)
	}

	// 发送设备离线状态
	err = h.platform.SendDeviceStatus(req.DeviceID, "0")
	if err != nil {
		h.logger.WithError(err).Error("发送设备离线状态失败")
		return err
	}

	return nil
}

// handleNotification 处理通知请求
func (h *HTTPHandler) handleNotification(req *handler.NotificationRequest) error {
	h.logger.WithFields(logrus.Fields{
		"message_type": req.MessageType,
		"message":      req.Message,
	}).Info("收到通知请求")

	// 解析消息内容
	var msgData map[string]interface{}
	if err := json.Unmarshal([]byte(req.Message), &msgData); err != nil {
		h.logger.WithError(err).Error("解析通知消息失败")
		return err
	}

	// 处理不同类型的通知
	switch req.MessageType {
	case "1": // 服务配置修改
		h.logger.Info("处理服务配置修改通知")
		// TODO: 实现服务配置修改逻辑
	case "2": // 设备配置修改
		h.logger.Info("处理设备配置修改通知")
		// TODO: 实现设备配置修改逻辑
	default:
		h.logger.Warnf("未知的通知类型: %s", req.MessageType)
	}

	return nil
}

// handleGetDeviceList 处理获取设备列表请求
func (h *HTTPHandler) handleGetDeviceList(req *handler.GetDeviceListRequest) (*handler.DeviceListResponse, error) {
	h.logger.WithFields(logrus.Fields{
		"voucher":            req.Voucher,
		"service_identifier": req.ServiceIdentifier,
		"page":               req.Page,
		"page_size":          req.PageSize,
	}).Info("收到获取设备列表请求")

	// 解析voucher, 其结构为：{"ServerURL":"http://127.0.0.1:8002/xiaozhi","Secret":"7cecb9b4-acde-4fb1-9c40-2a7f60e135ea","ThingsPanelApiKey":"sk_e6e72a3ef2aa2e7f8f15a9822a72c58bbc754aba4589df84d5d58a71c046c5fe","ThingsPanelApiURL":"http://thingspanel.local/api/v1"}
	var voucher formjson.Voucher
	if err := json.Unmarshal([]byte(req.Voucher), &voucher); err != nil {
		h.logger.WithError(err).Error("解析凭证失败")
		return nil, err
	}

	// 调用vourcher中的serverurl的/device/list接口, header中带上secret, 并将原始req中所有参数原封不动用post传递给/device/list接口
	requestData := map[string]interface{}{
		"voucher":            req.Voucher,
		"service_identifier": req.ServiceIdentifier,
		"page":               req.Page,
		"page_size":          req.PageSize,
	}
	requestBody, err := json.Marshal(requestData)
	if err != nil {
		h.logger.WithError(err).Error("序列化请求数据失败")
		return nil, err
	}

	// 发送POST请求
	httpReq, err := http.NewRequest("POST", voucher.ServerURL+"/device/list", bytes.NewBuffer(requestBody))
	if err != nil {
		h.logger.WithError(err).Error("创建请求失败")
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-token", voucher.Secret)

	// 将请求的request url, header, body写入日志
	h.logger.WithFields(logrus.Fields{
		"url":    httpReq.URL.String(),
		"header": httpReq.Header,
		"body":   string(requestBody),
	}).Info("发送第三方请求")

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		h.logger.WithError(err).Error("调用第三方接口失败")
		return nil, err
	}
	defer resp.Body.Close()

	// 读取响应体
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		h.logger.WithError(err).Error("读取响应体失败")
		return nil, err
	}

	// 将接口返回的信息写入日志
	h.logger.WithFields(logrus.Fields{
		"status_code": resp.StatusCode,
		"body":        string(bodyBytes),
	}).Info("第三方接口响应")

	// 解析响应
	var responseData struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			Total int `json:"total"`
			List  []struct {
				DeviceName   string `json:"device_name"`
				DeviceNumber string `json:"device_number"`
				Description  string `json:"description"`
			} `json:"list"`
		} `json:"data"`
	}
	if err := json.Unmarshal(bodyBytes, &responseData); err != nil {
		h.logger.WithError(err).Error("解析响应数据失败")
		return nil, err
	}

	// 组装DeviceListData
	deviceListData := handler.DeviceListData{
		List:  []handler.DeviceItem{},
		Total: responseData.Data.Total,
	}
	for _, device := range responseData.Data.List {
		deviceListData.List = append(deviceListData.List, handler.DeviceItem{
			DeviceName:   device.DeviceName,
			DeviceNumber: device.DeviceNumber,
			Description:  device.Description,
		})
	}

	rsp := handler.DeviceListResponse{
		Code:    200,
		Message: "获取成功",
		Data:    deviceListData,
	}

	// 将最终的rsp写入日志
	h.logger.WithFields(logrus.Fields{
		"code":    rsp.Code,
		"message": rsp.Message,
		"data":    rsp.Data,
	}).Info("接口响应")

	return &rsp, nil
}

// handleGetDeviceInfo 处理获取设备详细请求
func (h *HTTPHandler) handleGetDeviceInfo(req *handler.GetDeviceInfoRequest) (*handler.GetDeviceInfoResponse, error) {
	h.logger.WithFields(logrus.Fields{
		"device_code": req.Key,
		"voucher":     req.Voucher,
		"raw_request": fmt.Sprintf("%+v", req),
	}).Info("收到获取设备详细请求")

	// 检查请求参数
	if req.Key == "" {
		h.logger.Error("设备编码为空")
		return nil, fmt.Errorf("设备编码不能为空")
	}

	if req.Voucher == "" {
		h.logger.Error("凭证为空")
		return nil, fmt.Errorf("凭证不能为空")
	}

	// 解析Voucher
	var voucher struct {
		ServerURL         string `json:"ServerURL"`
		Secret            string `json:"Secret"`
		AgentId           string `json:"AgentId"`
		ThingsPanelApiKey string `json:"ThingsPanelApiKey"`
	}
	if err := json.Unmarshal([]byte(req.Voucher), &voucher); err != nil {
		h.logger.WithError(err).Error("解析Voucher失败")
		return nil, err
	}

	// 检查Voucher中的必要字段
	if voucher.ServerURL == "" || voucher.Secret == "" || voucher.AgentId == "" || voucher.ThingsPanelApiKey == "" {
		h.logger.Error("Voucher中缺少必要字段")
		return nil, fmt.Errorf("Voucher中缺少必要字段")
	}

	// 准备请求数据
	requestData := map[string]string{
		"secret":           voucher.Secret,
		"agent_id":         voucher.AgentId,
		"external_api_key": voucher.ThingsPanelApiKey,
		"device_code":      req.Key,
	}
	requestBody, err := json.Marshal(requestData)
	if err != nil {
		h.logger.WithError(err).Error("序列化请求数据失败")
		return nil, err
	}

	// 发送POST请求
	resp, err := http.Post(voucher.ServerURL+"/device/bind", "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		h.logger.WithError(err).Error("调用第三方接口失败")
		return nil, err
	}
	defer resp.Body.Close()

	// 读取响应体
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		h.logger.WithError(err).Error("读取响应体失败")
		return nil, err
	}

	// 输出响应体日志
	h.logger.WithFields(logrus.Fields{
		"status_code": resp.StatusCode,
		"body":        string(bodyBytes),
	}).Info("第三方接口响应")

	// 解析响应
	var responseData struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			DeviceName        string `json:"device_name"`
			DeviceNumber      string `json:"device_number"`
			DeviceDescription string `json:"device_description"`
		} `json:"data"`
	}
	if err := json.Unmarshal(bodyBytes, &responseData); err != nil {
		h.logger.WithError(err).Error("解析响应数据失败")
		return nil, err
	}

	// 检查响应码
	if responseData.Code != 0 {
		h.logger.WithFields(logrus.Fields{
			"code": responseData.Code,
			"msg":  responseData.Msg,
		}).Error("第三方接口返回错误")
		return nil, fmt.Errorf("第三方接口错误: %s", responseData.Msg)
	}

	// 组装DeviceItem
	deviceItem := handler.DeviceItem{
		DeviceName:   responseData.Data.DeviceName,
		DeviceNumber: responseData.Data.DeviceNumber,
		Description:  responseData.Data.DeviceDescription,
	}

	rsp := handler.GetDeviceInfoResponse{
		Code:    200,
		Message: "获取成功",
		Data:    deviceItem,
	}

	return &rsp, nil
}
