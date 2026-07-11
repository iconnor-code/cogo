# Repository Guidelines

## 项目结构与模块组织

本仓库是 Go 微服务基础组件库（`github.com/iconnor-code/cogo`）。核心接口位于 `core/`，配置、日志、服务、注册、发现和请求上下文等实现位于 `core/impl/`。外部依赖客户端封装在 `client/`，客户端选项在 `clientopt/`，gRPC unary 拦截器在 `interceptor/`，通用包在 `pkg/`，杂项工具在 `utils/`，错误模型在 `cerrs/`。项目文档维护在 `docs/`。

测试文件与被测代码放在同一包目录下，命名为 `*_test.go`。

## 构建、测试与开发命令

- `go test ./...`：运行所有包的完整测试。
- `go test ./pkg/token`：迭代时只运行单个包的测试。
- `go test -run TestName ./path/to/package`：运行指定测试用例。
- `go build ./...`：编译所有包并发现构建错误。
- `go mod tidy`：依赖变更后更新 module 元数据。

## 代码风格与命名约定

使用标准 Go 格式；完成前对改动的 Go 文件运行 `gofmt`。包名保持简短、小写，并与目录名一致。导出标识符使用 PascalCase；作为公共 API 时应补充注释。优先使用小接口和构造函数，并保持与 `core`、`core/impl` 中的既有模式一致。

处理聚焦问题时避免大范围重构。文档更新应放在 `docs/`，README 内容要与实际行为一致。

## 基础库边界

`cogo` 只承载跨服务基础设施能力，例如配置、日志、服务生命周期、注册发现、拦截器、通用客户端、通用错误模型和通用工具。不要引入账号、博客、评论等业务语义，也不要放业务 DTO、业务错误码或具体服务调用流程。

只有当一个能力至少被两个服务稳定复用，且可以脱离业务服务独立测试时，才考虑抽入 `cogo`。单个服务专用逻辑应保留在对应服务内。

## 测试规范

使用 Go 内置 `testing` 包。测试函数命名为 `TestXxx`，并与被测代码保存在同一包目录。行为变更需要新增或更新聚焦测试，尤其是拦截器、token 处理、服务生命周期和错误转换。修改共享抽象或公共工具时，运行 `go test ./...`。

## 提交与 Pull Request 规范

近期提交多使用简短、动作导向的摘要，并偶尔使用 conventional 前缀，例如 `docs: add project documentation set` 或 `Improve service lifecycle and token validation`。提交信息应简洁说明变更；当有助于表达范围时，可使用 `docs:`、`fix:`、`test:` 等前缀。

Pull Request 应描述变更内容、列出已执行的验证、关联相关 issue，并明确说明配置或行为变化。只有涉及用户可见界面变化时才需要截图。

## Agent 专用说明

编辑前先检查相关包；除非任务明确要求破坏性变更，否则保持现有 API。不要修改无关文件，也不要更新生成文件或依赖元数据，除非这是完成当前任务所必需的。
