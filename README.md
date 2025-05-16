# 协议插件模板

这是一个用于开发ThingsPanel协议插件的框架模板，提供了完整的插件开发基础架构，可以帮助开发者快速构建自定义协议插件。

## 特性

- 内置日志系统，支持文件轮转
- MQTT客户端集成
- 设备管理和缓存机制
- 表单配置管理
- HTTP服务支持
- 优雅的错误处理
- 配置文件管理

## 目录结构

```text
.
├── cmd/                    # 主程序入口
│   └── main.go            # 主程序
├── configs/               # 配置文件目录
│   └── config.yaml        # 主配置文件
├── internal/              # 内部包
│   ├── config/           # 配置结构定义
│   ├── form_json/        # 表单JSON定义
│   ├── handler/          # HTTP处理器
│   ├── pkg/              # 通用包
│   │   └── logger/       # 日志包
│   └── platform/         # 平台交互
├── examples/              # 示例代码
└── go.mod                # Go模块文件
```

## 核心组件说明

### 1. 配置管理 (internal/config)

- 定义了插件所需的各种配置结构
- 支持服务器配置、平台配置和日志配置
- 使用YAML格式配置文件

### 2. HTTP处理器 (internal/handler)

- 处理各种HTTP请求
- 实现了表单配置、设备断开连接、通知等处理函数
- 支持自定义处理逻辑

### 3. 日志系统 (internal/pkg/logger)

- 基于logrus的日志系统
- 支持日志级别控制
- 支持日志文件轮转
- 支持控制台彩色输出

### 4. 平台客户端 (internal/platform)

- 管理与ThingsPanel平台的通信
- 提供设备管理和缓存机制
- 处理遥测数据发送
- 管理设备状态和心跳

## 规范

- 官方插件开发说明文档

<http://thingspanel.io/zh-Hans/docs/system-development/eveloping-plug-in/customProtocol>

## 表单规范

表单 JSON 结构规范

该文档描述了用于生成前端表单的 JSON 结构的规范。它确定了必须和可选字段，以及它们的预期值。

1. 总体结构
表单由一个数组构成，每个数组元素都代表一个表单元素。表单元素可以是各种类型，如输入框或表格。

    ```text
    [
        { /* 表单元素1 */ },
        { /* 表单元素2 */ },
        // ...
    ]
    ```

2. 字段定义

| 字段名称    | 必选/可选               | 数据类型 | 描述                                                                           | 示例或备注                            |
| ----------- | ----------------------- | -------- | ------------------------------------------------------------------------------ | ------------------------------------- |
| dataKey     | 必填                    | 字符串   | 用于唯一标识表单元素的键。                                                     | "temp", "table1"                      |
| label       | 必填                    | 字符串   | 显示为表单元素标签的文本。                                                     | "读取策略(秒)", "属性列表"            |
| placeholder | 可选                    | 字符串   | 显示在表单元素中作为提示的文本。                                               | "请输入时间间隔，单位s"               |
| type        | 必填                    | 字符串   | 表单元素的类型。目前支持的类型有："input" 和 "table"。                         | "input", "table"                      |
| validate    | 可选                    | 对象     | 包含表单验证规则的对象。                                                       | 见 validate 字段的详细描述            |
| └─message   | 必填                    | 字符串   | 当验证失败时显示的错误消息。                                                   | "读取策略不能为空"                    |
| └─required  | 可选                    | 布尔值   | 指定字段是否是必填项。                                                         | true, false                           |
| └─rules     | 可选                    | 字符串   | 用于验证字段值的正则表达式规则。                                               | "/^\d{1,}$/" — 值必须是一个或多个数字 |
| └─type      | 可选                    | 字符串   | 用于指定验证的类型，例如，"number" 表示字段值应为数字。                        | "number"                              |
| array       | 只在 "table" 类型中可用 | 数组     | 包含表格列定义的数组。每一个列定义都是一个表单元素对象，它有相同的结构和属性。 | 见 array 字段的详细描述               |

注意：

1. `validate` 字段和其子字段（`message`, `required`, `rules`, `type`）是一个嵌套的结构，它们定义了表单元素的验证规则。
2. `array` 字段只适用于类型为 "table" 的表单元素，并包含一个嵌套的表单元素对象数组，用于定义表格的列。
3. 示例
查看附录
4. 开发注意事项
提供开发人员注意事项和最佳实践，包括但不限于:
●保证 dataKey 的唯一性。
●为每个字段提供合适的 placeholder 来指导用户输入。
●使用合适的正则表达式进行输入验证。

### 附录

示例

```text
[
    {
  "dataKey": "temp",
  "label": "读取策略(秒)",
  "placeholder": "请输入时间间隔，单位s",
  "type": "input",
  "validate": {
   "message": "读取策略不能为空",
   "required": true,
   "rules": "/^\\d{1,}$/",
   "type": "number"
  }
 },
 {
  "type": "table",
  "label": "属性列表",
       "dataKey": "table1",
  "array": [
   {
    "dataKey": "Interval",
    "label": "读取策略(秒)",
    "placeholder": "请输入时间间隔，单位s",
    "type": "input",
    "validate": {
     "message": "读取策略不能为空",
     "required": true,
     "rules": "/^\\d{1,}$/",
     "type": "number"
    }
   }
  ]
 }
]
```

表单填写后生成的数据样例

```json
{
 "attribute1":0,
 "attribute2": "",
 "table1": [
  {
   "attribute1":0,
      "attribute2": ""
  },
  {
   "attribute1":0,
        "attribute2": ""
  }
 ]
}
```
