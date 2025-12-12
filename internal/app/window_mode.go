package app

// WindowModeService 窗口模式服务
type WindowModeService struct {
	mode string
}

// NewWindowModeService 创建窗口模式服务
func NewWindowModeService(mode string) *WindowModeService {
	return &WindowModeService{mode: mode}
}

// GetMode 获取当前窗口模式
func (w *WindowModeService) GetMode() string {
	return w.mode
}
