package ui

import (
	"context"
	"fmt"
	"fridare-gui/internal/config"
	"fridare-gui/internal/core"
	"path/filepath"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// FixedWidthEntry 固定宽度的Entry组件
type FixedWidthEntry struct {
	widget.Entry
	fixedWidth float32
}

// NewFixedWidthEntry 创建固定宽度的Entry
func NewFixedWidthEntry(width float32) *FixedWidthEntry {
	entry := &FixedWidthEntry{
		fixedWidth: width,
	}
	entry.ExtendBaseWidget(entry)
	return entry
}

// MinSize 返回固定的最小尺寸
func (e *FixedWidthEntry) MinSize() fyne.Size {
	return fyne.NewSize(e.fixedWidth, 35)
}

// FixedWidthSelect 固定宽度的Select组件
type FixedWidthSelect struct {
	widget.Select
	fixedWidth float32
}

// NewFixedWidthSelect 创建固定宽度的Select
func NewFixedWidthSelect(options []string, width float32) *FixedWidthSelect {
	sel := &FixedWidthSelect{
		fixedWidth: width,
	}
	sel.Select.Options = options
	sel.ExtendBaseWidget(sel)
	return sel
}

// MinSize 返回固定的最小尺寸
func (s *FixedWidthSelect) MinSize() fyne.Size {
	return fyne.NewSize(s.fixedWidth, 35)
}

// 创建标准字体大小的标签
func newStandardLabel(text string) *widget.Label {
	label := widget.NewLabel(text)
	// 设置标准文本样式
	label.TextStyle = fyne.TextStyle{}
	return label
}

// 创建小号字体的标签（用于详细信息）
func newSmallLabel(text string) *widget.Label {
	label := widget.NewLabel(text)
	label.TextStyle = fyne.TextStyle{}
	return label
}

// 创建粗体标签（用于文件名）
func newBoldLabel(text string) *widget.Label {
	label := widget.NewLabel(text)
	label.TextStyle = fyne.TextStyle{Bold: true}
	return label
}

// AssetInfo 资源信息
type AssetInfo struct {
	Asset      core.Asset
	Version    string
	Platform   string
	FileType   string
	SHA256     string
	Size       string
	UploadTime string
	Selected   bool
	// 下载状态相关
	IsDownloading bool
	IsPaused      bool
	Progress      float64 // 0.0 - 1.0
	Speed         string
	Downloaded    string
	Status        string // "等待", "下载中", "暂停", "完成", "失败"
}

// DownloadTab 下载标签页
type DownloadTab struct {
	app          fyne.App
	config       *config.Config
	updateStatus StatusUpdater

	// UI 组件
	content        *fyne.Container
	versionSelect  *widget.Select
	customVersion  *FixedWidthEntry
	platformSelect *FixedWidthSelect
	assetList      *widget.List
	filterEntry    *FixedWidthEntry

	// 工具栏控件
	toolbarSelectAll *widget.Button
	toolbarInvertSel *widget.Button
	toolbarDownload  *widget.Button
	toolbarCancelSel *widget.Button
	toolbarStart     *widget.Button
	toolbarPause     *widget.Button
	toolbarResume    *widget.Button
	toolbarStop      *widget.Button

	// 底部进度可视化区域
	progressVisual *fyne.Container

	// 下载控制
	downloadPaused  bool
	downloadStopped bool
	activeDownloads map[int]*DownloadTask // 活跃下载任务

	// 资源数据
	currentAssets    []AssetInfo
	filteredAssets   []AssetInfo
	selectedAssets   map[int]bool
	currentSelection int // 当前选中的文件索引

	// 业务组件
	fridaClient *core.FridaClient
	versions    []core.FridaVersion
}

// DownloadTask 下载任务
type DownloadTask struct {
	AssetIndex int
	Progress   float64
	Speed      string
	Downloaded string
	Status     string
	IsPaused   bool
	Context    context.Context
	CancelFunc context.CancelFunc
}

// NewDownloadTab 创建下载标签页
func NewDownloadTab(app fyne.App, cfg *config.Config, statusUpdater StatusUpdater) *DownloadTab {
	dt := &DownloadTab{
		app:              app,
		config:           cfg,
		updateStatus:     statusUpdater,
		activeDownloads:  make(map[int]*DownloadTask),
		currentSelection: -1, // 初始化为无选择
	}

	// 创建 Frida 客户端 - 为下载设置更长的超时时间
	downloadTimeout := time.Duration(cfg.Timeout*3) * time.Second // 下载超时设为普通超时的3倍
	if downloadTimeout < 120*time.Second {
		downloadTimeout = 120 * time.Second // 至少2分钟
	}
	dt.fridaClient = core.NewFridaClient(cfg.Proxy, downloadTimeout)

	dt.setupUI()
	dt.loadVersions()

	return dt
}

// setupUI 设置UI
func (dt *DownloadTab) setupUI() {
	// 初始化选中状态
	dt.selectedAssets = make(map[int]bool)

	// 创建顶部控制区域（版本选择 + 工具栏）
	topSection := dt.createTopSection()

	// 创建文件列表区域
	listSection := dt.createFileList()

	// 创建状态栏
	statusSection := dt.createStatusBar()

	// 主要布局：顶部控制 + 文件列表 + 状态栏
	dt.content = container.NewBorder(
		topSection,    // 顶部
		statusSection, // 底部
		nil, nil,      // 左右
		listSection, // 中心
	)
}

// createTopSection 创建顶部控制区域（包含版本选择和工具栏）
func (dt *DownloadTab) createTopSection() *fyne.Container {
	// 第一行：版本和平台选择
	dt.versionSelect = widget.NewSelect([]string{"加载中..."}, nil)
	dt.versionSelect.OnChanged = dt.onVersionChanged

	// 使用固定宽度的自定义版本输入框
	dt.customVersion = NewFixedWidthEntry(150)
	dt.customVersion.SetPlaceHolder("或手动输入版本号")

	refreshBtn := widget.NewButton("刷新", func() {
		dt.loadVersions()
	})

	// 平台选择 - 使用固定宽度的自定义Select
	platformOptions := make([]string, len(core.SupportedPlatforms)+1)
	platformOptions[0] = "All" // 添加 All 选项作为第一个选项
	for i, platform := range core.SupportedPlatforms {
		platformOptions[i+1] = platform.Name
	}
	dt.platformSelect = NewFixedWidthSelect(platformOptions, 120)
	dt.platformSelect.SetSelected("All") // 默认选择 All
	dt.platformSelect.OnChanged = dt.onPlatformChanged

	// 过滤器 - 使用固定宽度的自定义Entry
	dt.filterEntry = NewFixedWidthEntry(120)
	dt.filterEntry.SetPlaceHolder("过滤文件名...")
	dt.filterEntry.OnChanged = dt.onFilterChanged

	// 第一行：版本和平台控制 - 使用分组布局确保清晰的结构和间距
	versionSection := container.NewHBox(
		newStandardLabel("版本:"),
		dt.versionSelect,
		dt.customVersion,
		refreshBtn,
	)

	platformSection := container.NewHBox(
		newStandardLabel("平台:"),
		dt.platformSelect,
	)

	filterSection := container.NewHBox(
		newStandardLabel("过滤:"),
		dt.filterEntry,
	)

	// 第一行：版本和平台控制 - 使用分组布局确保清晰的结构和间距
	firstRow := container.NewHBox(
		versionSection,
		widget.NewSeparator(),
		platformSection,
		widget.NewSeparator(),
		filterSection,
	)

	// 第二行：文件操作工具栏
	dt.toolbarSelectAll = widget.NewButton("全选", dt.selectAll)
	dt.toolbarSelectAll.Resize(fyne.NewSize(60, 35))

	dt.toolbarInvertSel = widget.NewButton("反选", dt.invertSelection)
	dt.toolbarInvertSel.Resize(fyne.NewSize(60, 35))

	dt.toolbarDownload = widget.NewButton("下载选中", dt.startBatchDownload)
	dt.toolbarDownload.Importance = widget.HighImportance
	dt.toolbarDownload.Resize(fyne.NewSize(100, 35))
	dt.toolbarDownload.Disable()

	dt.toolbarCancelSel = widget.NewButton("取消选中", dt.clearSelection)
	dt.toolbarCancelSel.Resize(fyne.NewSize(80, 35))

	// 下载控制按钮
	dt.toolbarStart = widget.NewButton("开始", dt.startSelectedDownloads)
	dt.toolbarStart.Resize(fyne.NewSize(60, 35))
	dt.toolbarStart.Disable()

	dt.toolbarPause = widget.NewButton("暂停", dt.pauseSelectedDownloads)
	dt.toolbarPause.Resize(fyne.NewSize(60, 35))
	dt.toolbarPause.Disable()

	dt.toolbarResume = widget.NewButton("继续", dt.resumeSelectedDownloads)
	dt.toolbarResume.Resize(fyne.NewSize(60, 35))
	dt.toolbarResume.Disable()

	dt.toolbarStop = widget.NewButton("停止", dt.stopSelectedDownloads)
	dt.toolbarStop.Resize(fyne.NewSize(60, 35))
	dt.toolbarStop.Disable()

	secondRow := container.NewHBox(
		dt.toolbarSelectAll,
		dt.toolbarInvertSel,
		dt.toolbarDownload,
		dt.toolbarCancelSel,
		widget.NewSeparator(),
		dt.toolbarStart,
		dt.toolbarPause,
		dt.toolbarResume,
		dt.toolbarStop,
	)

	// 垂直组合两行
	return container.NewVBox(
		firstRow,
		widget.NewSeparator(),
		secondRow,
	)
}

// createStatusBar 创建状态栏
func (dt *DownloadTab) createStatusBar() *fyne.Container {
	return container.NewHBox(
		newStandardLabel("就绪"),
	)
}

// createFileList 创建文件列表
func (dt *DownloadTab) createFileList() *container.Scroll {
	// 资源列表
	dt.assetList = widget.NewList(
		func() int {
			return len(dt.filteredAssets)
		},
		func() fyne.CanvasObject {
			check := widget.NewCheck("", nil)
			check.Resize(fyne.NewSize(20, 20))

			nameLabel := newBoldLabel("")
			nameLabel.Wrapping = fyne.TextWrapOff

			infoLabel := newSmallLabel("")

			// 下载进度条 - 紧凑型
			progressBar := widget.NewProgressBar()
			progressBar.SetValue(0)
			progressBar.Hide()
			progressBar.Resize(fyne.NewSize(120, 18)) // 限制进度条宽度

			// 控制按钮 - 紧凑型
			pauseBtn := widget.NewButton("⏸", nil)
			pauseBtn.Resize(fyne.NewSize(22, 18))
			pauseBtn.Hide()

			stopBtn := widget.NewButton("⏹", nil)
			stopBtn.Resize(fyne.NewSize(22, 18))
			stopBtn.Hide()

			// 状态标签 - 紧凑型
			statusLabel := newSmallLabel("")
			statusLabel.Hide()

			// 创建文件名和控制区域的容器
			fileNameContainer := container.NewHBox(nameLabel)

			controlsContainer := container.NewHBox(
				progressBar,
				statusLabel,
				pauseBtn,
				stopBtn,
			)
			controlsContainer.Hide()

			// 第一行：文件名 + 控制区域 - 使用HBox而不是Border
			firstRow := container.NewHBox(
				fileNameContainer,
				widget.NewSeparator(), // 分隔符
				controlsContainer,
			)

			// 整个内容区域
			content := container.NewVBox(
				firstRow,
				infoLabel,
			)

			// 最终项目容器
			item := container.NewBorder(
				nil, nil, // top, bottom
				check, nil, // left, right
				content, // center
			)

			return item
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id >= len(dt.filteredAssets) {
				return
			}

			borderContainer := obj.(*fyne.Container)
			if len(borderContainer.Objects) < 2 {
				return
			}

			// 找到各个控件
			var check *widget.Check
			var content *fyne.Container

			for _, child := range borderContainer.Objects {
				if c, ok := child.(*widget.Check); ok {
					check = c
				} else if cont, ok := child.(*fyne.Container); ok {
					content = cont
				}
			}

			if check == nil || content == nil || len(content.Objects) < 2 {
				return
			}

			// 获取布局结构
			firstRow := content.Objects[0].(*fyne.Container) // 第一行
			infoLabel := content.Objects[1].(*widget.Label)  // 详细信息

			// 从第一行获取文件名和控制区域 (HBox: nameContainer, separator, controlsContainer)
			fileNameContainer := firstRow.Objects[0].(*fyne.Container) // 文件名容器
			controlsContainer := firstRow.Objects[2].(*fyne.Container) // 控制区域 (跳过分隔符)

			// 获取具体控件
			nameLabel := fileNameContainer.Objects[0].(*widget.Label)
			progressBar := controlsContainer.Objects[0].(*widget.ProgressBar)
			statusLabel := controlsContainer.Objects[1].(*widget.Label)
			pauseBtn := controlsContainer.Objects[2].(*widget.Button)
			stopBtn := controlsContainer.Objects[3].(*widget.Button)

			asset := dt.filteredAssets[id]
			check.SetChecked(dt.selectedAssets[id])
			check.OnChanged = func(checked bool) {
				dt.selectedAssets[id] = checked
				dt.updateDownloadButton()
				dt.updateToolbarDownloadButtons() // 更新工具栏按钮状态
			}

			// 设置文件名
			nameLabel.SetText(asset.Asset.Name)

			// 设置详细信息
			info := fmt.Sprintf("平台: %s | 类型: %s | 大小: %s | 时间: %s",
				asset.Platform,
				asset.FileType,
				asset.Size,
				asset.UploadTime)
			infoLabel.SetText(info)

			// 根据下载状态设置UI
			if asset.IsDownloading || asset.Status == "完成" || asset.Status == "失败" {
				// 显示控制区域
				controlsContainer.Show()
				progressBar.Show()
				progressBar.SetValue(asset.Progress)

				if asset.IsDownloading {
					pauseBtn.Show()
					stopBtn.Show()
					statusLabel.Show()
					statusLabel.SetText(fmt.Sprintf("%.1f%% - %s", asset.Progress*100, asset.Speed))

					// 根据暂停状态设置按钮文本
					if asset.IsPaused {
						pauseBtn.SetText("▶") // 继续
					} else {
						pauseBtn.SetText("⏸") // 暂停
					}

					pauseBtn.OnTapped = func() {
						dt.toggleDownloadPause(id)
						dt.updateToolbarDownloadButtons() // 更新工具栏按钮状态
					}
					stopBtn.OnTapped = func() {
						dt.stopAssetDownload(id)
						dt.updateToolbarDownloadButtons() // 更新工具栏按钮状态
					}
				} else if asset.Status == "完成" {
					pauseBtn.Hide()
					stopBtn.Hide()
					statusLabel.Show()
					statusLabel.SetText("完成")
				} else if asset.Status == "失败" {
					pauseBtn.Hide()
					stopBtn.Hide()
					statusLabel.Show()
					statusLabel.SetText("失败")
				}
			} else {
				// 隐藏控制区域，节约空间
				controlsContainer.Hide()
			}
		},
	)

	// 添加列表选择变化监听器
	dt.assetList.OnSelected = func(id widget.ListItemID) {
		// 当选择列表项时，更新当前选择索引和工具栏按钮状态
		dt.currentSelection = int(id)
		dt.updateToolbarDownloadButtons()
	}

	dt.assetList.OnUnselected = func(id widget.ListItemID) {
		// 当取消选择列表项时，重置选择索引并更新工具栏按钮状态
		dt.currentSelection = -1
		dt.updateToolbarDownloadButtons()
	}

	// 给列表添加滚动
	return container.NewScroll(dt.assetList)
}

// onVersionChanged 版本选择改变事件
func (dt *DownloadTab) onVersionChanged(selected string) {
	dt.loadAssetsForVersion(selected)
}

// onPlatformChanged 平台选择改变事件
func (dt *DownloadTab) onPlatformChanged(selected string) {
	dt.filterAssets()
}

// onFilterChanged 过滤器改变事件
func (dt *DownloadTab) onFilterChanged(text string) {
	dt.filterAssets()
}

// clearSelection 清除选择
func (dt *DownloadTab) clearSelection() {
	for i := range dt.filteredAssets {
		dt.selectedAssets[i] = false
	}
	fyne.Do(func() {
		dt.assetList.Refresh()
	})
	dt.updateToolbarButtons()
}

// updateToolbarButtons 更新工具栏按钮状态
func (dt *DownloadTab) updateToolbarButtons() {
	hasSelected := false
	for _, selected := range dt.selectedAssets {
		if selected {
			hasSelected = true
			break
		}
	}

	if hasSelected {
		dt.toolbarDownload.Enable()
	} else {
		dt.toolbarDownload.Disable()
	}
}

// toggleDownloadPause 切换下载暂停状态
func (dt *DownloadTab) toggleDownloadPause(assetIndex int) {
	if task, exists := dt.activeDownloads[assetIndex]; exists {
		task.IsPaused = !task.IsPaused
		if assetIndex < len(dt.filteredAssets) {
			if task.IsPaused {
				dt.filteredAssets[assetIndex].Status = "暂停"
				dt.filteredAssets[assetIndex].IsPaused = true
			} else {
				dt.filteredAssets[assetIndex].Status = "下载中"
				dt.filteredAssets[assetIndex].IsPaused = false
			}
			fyne.Do(func() {
				dt.assetList.Refresh()
			})
		}
	}
}

// stopAssetDownload 停止单个资源下载
func (dt *DownloadTab) stopAssetDownload(assetIndex int) {
	if task, exists := dt.activeDownloads[assetIndex]; exists {
		if task.CancelFunc != nil {
			task.CancelFunc()
		}
		delete(dt.activeDownloads, assetIndex)

		if assetIndex < len(dt.filteredAssets) {
			dt.filteredAssets[assetIndex].Status = "等待"
			dt.filteredAssets[assetIndex].IsDownloading = false
			dt.filteredAssets[assetIndex].IsPaused = false
			dt.filteredAssets[assetIndex].Progress = 0
			fyne.Do(func() {
				dt.assetList.Refresh()
			})
		}
	}
}

// selectAll 全选
func (dt *DownloadTab) selectAll() {
	for i := range dt.filteredAssets {
		dt.selectedAssets[i] = true
	}
	fyne.Do(func() {
		dt.assetList.Refresh()
	})
	dt.updateDownloadButton()
}

// invertSelection 反选
func (dt *DownloadTab) invertSelection() {
	for i := range dt.filteredAssets {
		dt.selectedAssets[i] = !dt.selectedAssets[i]
	}
	fyne.Do(func() {
		dt.assetList.Refresh()
	})
	dt.updateDownloadButton()
}

// updateDownloadButton 更新下载按钮状态（保持兼容性）
func (dt *DownloadTab) updateDownloadButton() {
	dt.updateToolbarButtons()
}

// loadAssetsForVersion 为指定版本加载资源
func (dt *DownloadTab) loadAssetsForVersion(versionName string) {
	var selectedVersion *core.FridaVersion
	for _, version := range dt.versions {
		if version.Version == versionName {
			selectedVersion = &version
			break
		}
	}

	if selectedVersion == nil {
		dt.currentAssets = []AssetInfo{}
		dt.filterAssets()
		return
	}

	// 转换资源为AssetInfo
	dt.currentAssets = make([]AssetInfo, 0, len(selectedVersion.Assets))
	for _, asset := range selectedVersion.Assets {
		// 解析文件名以确定平台和文件类型
		platform, fileType := dt.parseAssetInfo(asset.Name)

		assetInfo := AssetInfo{
			Asset:         asset,
			Version:       selectedVersion.Version,
			Platform:      platform,
			FileType:      fileType,
			SHA256:        "计算中...", // 这里可以后续从GitHub API获取
			Size:          core.FormatSize(asset.Size),
			UploadTime:    selectedVersion.Published.Format("2006-01-02"),
			Selected:      false,
			IsDownloading: false,
			IsPaused:      false,
			Progress:      0.0,
			Speed:         "",
			Downloaded:    "",
			Status:        "等待",
		}
		dt.currentAssets = append(dt.currentAssets, assetInfo)
	}

	dt.filterAssets()
}

// parseAssetInfo 解析资源文件名获取平台和文件类型信息
func (dt *DownloadTab) parseAssetInfo(filename string) (platform, fileType string) {
	// 简单的文件名解析逻辑
	lower := strings.ToLower(filename)

	// 确定文件类型
	if strings.Contains(lower, "server") {
		fileType = "frida-server"
	} else if strings.Contains(lower, "gadget") {
		fileType = "frida-gadget"
	} else if strings.Contains(lower, "tools") || strings.Contains(lower, ".whl") {
		fileType = "frida-tools"
	} else {
		fileType = "其他"
	}

	// 确定平台
	if strings.Contains(lower, "android") {
		if strings.Contains(lower, "arm64") {
			platform = "Android ARM64"
		} else if strings.Contains(lower, "arm") {
			platform = "Android ARM"
		} else if strings.Contains(lower, "x86_64") {
			platform = "Android x86_64"
		} else if strings.Contains(lower, "x86") {
			platform = "Android x86"
		} else {
			platform = "Android"
		}
	} else if strings.Contains(lower, "ios") {
		platform = "iOS ARM64"
	} else if strings.Contains(lower, "windows") {
		if strings.Contains(lower, "x86_64") || strings.Contains(lower, "amd64") {
			platform = "Windows x64"
		} else {
			platform = "Windows x86"
		}
	} else if strings.Contains(lower, "linux") {
		if strings.Contains(lower, "arm64") {
			platform = "Linux ARM64"
		} else {
			platform = "Linux x64"
		}
	} else if strings.Contains(lower, "macos") || strings.Contains(lower, "darwin") {
		if strings.Contains(lower, "arm64") {
			platform = "macOS ARM64"
		} else {
			platform = "macOS x64"
		}
	} else {
		platform = "通用"
	}

	return platform, fileType
}

// filterAssets 过滤资源
func (dt *DownloadTab) filterAssets() {
	if dt.assetList == nil {
		return // 如果列表还没初始化，直接返回
	}

	platformFilter := dt.platformSelect.Selected
	textFilter := strings.ToLower(dt.filterEntry.Text)

	dt.filteredAssets = []AssetInfo{}
	dt.selectedAssets = make(map[int]bool)
	dt.currentSelection = -1 // 重置当前选择

	filteredIndex := 0
	for _, asset := range dt.currentAssets {
		// 平台过滤 - 如果选择 "All" 则不进行平台过滤
		if platformFilter != "" && platformFilter != "All" && asset.Platform != platformFilter && asset.Platform != "通用" {
			continue
		}

		// 文本过滤
		if textFilter != "" && !strings.Contains(strings.ToLower(asset.Asset.Name), textFilter) {
			continue
		}

		dt.filteredAssets = append(dt.filteredAssets, asset)
		dt.selectedAssets[filteredIndex] = false
		filteredIndex++
	}

	fyne.Do(func() {
		dt.assetList.Refresh()
	})
	dt.updateDownloadButton()
	dt.updateToolbarDownloadButtons()
}

// startBatchDownload 开始批量下载
func (dt *DownloadTab) startBatchDownload() {
	// 收集选中的资源
	var selectedAssets []int // 存储选中资源的索引
	for i, selected := range dt.selectedAssets {
		if selected && i < len(dt.filteredAssets) {
			selectedAssets = append(selectedAssets, i)
		}
	}

	if len(selectedAssets) == 0 {
		dt.updateStatus("请选择要下载的文件")
		return
	}

	// 禁用下载按钮
	dt.toolbarDownload.Disable()
	dt.updateStatus(fmt.Sprintf("开始批量下载 %d 个文件...", len(selectedAssets)))

	// 为每个选中的文件开始下载
	for _, assetIndex := range selectedAssets {
		dt.startSingleDownload(assetIndex)
	}
}

// startSingleDownload 开始单个文件下载
func (dt *DownloadTab) startSingleDownload(assetIndex int) {
	if assetIndex >= len(dt.filteredAssets) {
		return
	}

	// 更新资源状态
	dt.filteredAssets[assetIndex].Status = "下载中"
	dt.filteredAssets[assetIndex].IsDownloading = true
	dt.filteredAssets[assetIndex].Progress = 0.0

	// 刷新UI
	fyne.DoAndWait(func() {
		dt.assetList.Refresh()
	})

	// 创建带取消功能的上下文
	ctx, cancel := context.WithCancel(context.Background())

	// 创建下载任务
	task := &DownloadTask{
		AssetIndex: assetIndex,
		Progress:   0.0,
		Status:     "下载中",
		IsPaused:   false,
		Context:    ctx,
		CancelFunc: cancel,
	}
	dt.activeDownloads[assetIndex] = task

	go func() {
		defer cancel() // 确保goroutine结束时取消上下文

		asset := dt.filteredAssets[assetIndex]

		// 确保下载目录存在
		if err := dt.config.EnsureDownloadDir(); err != nil {
			dt.updateStatus(fmt.Sprintf("创建下载目录失败: %v", err))
			dt.filteredAssets[assetIndex].Status = "失败"
			dt.assetList.Refresh()
			return
		}

		// 构建下载路径
		downloadPath := filepath.Join(dt.config.DownloadDir, asset.Asset.Name)

		// 下载文件，带进度回调和暂停检查
		err := dt.fridaClient.DownloadFile(asset.Asset.DownloadURL, downloadPath, func(downloaded, total int64, speed float64) {
			// 检查上下文是否被取消
			select {
			case <-ctx.Done():
				return // 下载被取消
			default:
			}

			// 检查是否暂停
			if task, exists := dt.activeDownloads[assetIndex]; exists && task.IsPaused {
				// 暂停状态下不更新进度，但不退出
				return
			}

			// 更新进度
			if total > 0 {
				progress := float64(downloaded) / float64(total)
				dt.filteredAssets[assetIndex].Progress = progress
				dt.filteredAssets[assetIndex].Speed = core.FormatSpeed(speed)
				dt.filteredAssets[assetIndex].Downloaded = core.FormatSize(downloaded)

				// 刷新UI
				fyne.Do(func() {
					dt.assetList.Refresh()
				})
			}
		})

		// 下载完成处理
		if err != nil {
			if ctx.Err() == context.Canceled {
				dt.updateStatus(fmt.Sprintf("下载已停止: %s", asset.Asset.Name))
			} else {
				dt.updateStatus(fmt.Sprintf("下载失败 %s: %v", asset.Asset.Name, err))
			}
			dt.filteredAssets[assetIndex].Status = "失败"
		} else {
			dt.updateStatus(fmt.Sprintf("下载完成: %s", asset.Asset.Name))
			dt.filteredAssets[assetIndex].Status = "完成"
			dt.filteredAssets[assetIndex].Progress = 1.0

			// 更新最近使用
			dt.config.AddRecentVersion(asset.Version)
			dt.config.Save()
		}

		// 清理任务
		delete(dt.activeDownloads, assetIndex)
		dt.filteredAssets[assetIndex].IsDownloading = false
		fyne.DoAndWait(func() {
			dt.assetList.Refresh()
		})

		// 重新启用下载按钮
		dt.toolbarDownload.Enable()
	}()
}

// loadVersions 加载版本列表
func (dt *DownloadTab) loadVersions() {
	dt.updateStatus("正在获取版本列表...")
	dt.versionSelect.SetOptions([]string{"加载中..."})
	dt.versionSelect.Disable()

	go func() {
		versions, err := dt.fridaClient.GetVersions()

		// 在主线程中更新UI，使用fyne.Do避免线程警告
		fyne.Do(func() {
			if err != nil {
				dt.versionSelect.SetOptions([]string{"获取失败"})
				dt.versionSelect.Enable()
				dt.updateStatus(fmt.Sprintf("获取版本列表失败: %v", err))
				return
			}

			dt.versions = versions

			// 提取版本字符串
			versionStrings := make([]string, len(versions))
			for i, version := range versions {
				versionStrings[i] = version.Version
			}

			// 更新UI控件选项
			dt.versionSelect.SetOptions(versionStrings)
			dt.versionSelect.Enable()

			// 如果有版本，选择第一个
			if len(versionStrings) > 0 {
				dt.versionSelect.SetSelected(versionStrings[0])
				dt.loadAssetsForVersion(versionStrings[0])
			}

			dt.updateStatus(fmt.Sprintf("成功加载 %d 个版本", len(versions)))
		})
	}()
}

// Content 返回标签页内容
func (dt *DownloadTab) Content() *fyne.Container {
	return dt.content
}

// Refresh 刷新标签页
func (dt *DownloadTab) Refresh() {
	dt.loadVersions()
}

// startSelectedDownloads 开始当前选中的下载任务
func (dt *DownloadTab) startSelectedDownloads() {
	// 只处理当前高亮选中的文件
	if dt.currentSelection >= 0 && dt.currentSelection < len(dt.filteredAssets) {
		if !dt.filteredAssets[dt.currentSelection].IsDownloading {
			dt.startAssetDownload(dt.currentSelection)
		}
	}
	dt.updateToolbarDownloadButtons()
}

// startAssetDownload 开始指定资源的下载
func (dt *DownloadTab) startAssetDownload(assetIndex int) {
	// 检查索引有效性
	if assetIndex < 0 || assetIndex >= len(dt.filteredAssets) {
		dt.updateStatus("无效的文件索引")
		return
	}

	// 检查是否已经在下载中
	if dt.filteredAssets[assetIndex].IsDownloading {
		dt.updateStatus("该文件正在下载中")
		return
	}

	// 如果文件已完成下载，询问是否重新下载
	if dt.filteredAssets[assetIndex].Status == "完成" {
		dt.updateStatus("文件已下载完成，正在重新下载...")
		// 重置状态
		dt.filteredAssets[assetIndex].Status = "等待"
		dt.filteredAssets[assetIndex].Progress = 0.0
	}

	// 开始单个文件下载
	dt.startSingleDownload(assetIndex)

	// 更新状态消息
	asset := dt.filteredAssets[assetIndex]
	dt.updateStatus(fmt.Sprintf("开始下载: %s", asset.Asset.Name))
}

// pauseSelectedDownloads 暂停当前选中的下载任务
func (dt *DownloadTab) pauseSelectedDownloads() {
	// 只处理当前高亮选中的文件
	if dt.currentSelection >= 0 && dt.currentSelection < len(dt.filteredAssets) {
		if dt.filteredAssets[dt.currentSelection].IsDownloading && !dt.filteredAssets[dt.currentSelection].IsPaused {
			dt.toggleDownloadPause(dt.currentSelection)
		}
	}
	dt.updateToolbarDownloadButtons()
}

// resumeSelectedDownloads 继续当前选中的下载任务
func (dt *DownloadTab) resumeSelectedDownloads() {
	// 只处理当前高亮选中的文件
	if dt.currentSelection >= 0 && dt.currentSelection < len(dt.filteredAssets) {
		if dt.filteredAssets[dt.currentSelection].IsDownloading && dt.filteredAssets[dt.currentSelection].IsPaused {
			dt.toggleDownloadPause(dt.currentSelection)
		}
	}
	dt.updateToolbarDownloadButtons()
}

// stopSelectedDownloads 停止当前选中的下载任务
func (dt *DownloadTab) stopSelectedDownloads() {
	// 只处理当前高亮选中的文件
	if dt.currentSelection >= 0 && dt.currentSelection < len(dt.filteredAssets) {
		if dt.filteredAssets[dt.currentSelection].IsDownloading {
			dt.stopAssetDownload(dt.currentSelection)
		}
	}
	dt.updateToolbarDownloadButtons()
}

// updateToolbarDownloadButtons 更新工具栏下载控制按钮状态
func (dt *DownloadTab) updateToolbarDownloadButtons() {
	hasDownloading := false
	hasPaused := false
	hasActive := false
	hasSelected := false

	// 只检查当前高亮选中的文件状态
	if dt.currentSelection >= 0 && dt.currentSelection < len(dt.filteredAssets) {
		hasSelected = true
		asset := dt.filteredAssets[dt.currentSelection]
		if asset.IsDownloading {
			hasActive = true
			if asset.IsPaused {
				hasPaused = true
			} else {
				hasDownloading = true
			}
		}
	}

	// 开始按钮：只有在选中文件且没有下载时才启用
	if hasSelected && !hasActive {
		dt.toolbarStart.Enable()
		dt.toolbarStart.SetText("开始")
	} else {
		dt.toolbarStart.Disable()
		dt.toolbarStart.SetText("开始")
	}

	// 根据当前文件的下载状态启用/禁用按钮
	if hasDownloading {
		dt.toolbarPause.Enable()
		dt.toolbarPause.SetText("暂停")
	} else {
		dt.toolbarPause.Disable()
		dt.toolbarPause.SetText("暂停")
	}

	if hasPaused {
		dt.toolbarResume.Enable()
		dt.toolbarResume.SetText("继续")
	} else {
		dt.toolbarResume.Disable()
		dt.toolbarResume.SetText("继续")
	}

	if hasActive && !hasPaused {
		dt.toolbarStop.Enable()
		dt.toolbarStop.SetText("停止")
	} else {
		dt.toolbarStop.Disable()
		dt.toolbarStop.SetText("停止")
	}
}
