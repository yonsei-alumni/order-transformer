package main

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type largeTheme struct {
	fyne.Theme
}

func (t *largeTheme) Size(name fyne.ThemeSizeName) float32 {
	return t.Theme.Size(name) * 1.3
}

func main() {
	a := app.NewWithID("com.github.benelog.order-transformer")
	a.Settings().SetTheme(&largeTheme{Theme: theme.DefaultTheme()})
	w := a.NewWindow("주문 엑셀 파일 변환기")
	w.Resize(fyne.NewSize(700, 500))

	transformTab := buildTransformTab(w)
	validateTab := buildValidateTab(w)

	tabs := container.NewAppTabs(
		container.NewTabItem("변환", transformTab),
		container.NewTabItem("검증", validateTab),
	)

	w.SetContent(tabs)
	w.ShowAndRun()
}

func buildTransformTab(w fyne.Window) fyne.CanvasObject {
	message := widget.NewMultiLineEntry()
	message.SetPlaceHolder("변환 결과가 여기에 표시됩니다.")
	message.Wrapping = fyne.TextWrapWord

	btn := widget.NewButton("구매자 주문 엑셀 파일 선택", func() {
		fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil {
				message.SetText(fmt.Sprintf("[오류] %v", err))
				return
			}
			if reader == nil {
				return
			}
			reader.Close()

			sourcePath := reader.URI().Path()
			log := func(msg string) {
				message.SetText(message.Text + msg + "\n")
			}

			message.SetText("")
			log(fmt.Sprintf("[파일 선택] %s", filepath.Base(sourcePath)))
			log(fmt.Sprintf("[경로] %s", sourcePath))
			log("")
			log("엑셀 파일을 읽고 있습니다...")

			sourceOrders, err := ReadSourceOrders(sourcePath)
			if err != nil {
				log(fmt.Sprintf("[오류] %v", err))
				return
			}

			log(fmt.Sprintf("총 %d건의 주문 데이터를 읽었습니다.", len(sourceOrders)))

			cancelledCount := 0
			for _, order := range sourceOrders {
				if order.IsCancelled() {
					cancelledCount++
				}
			}
			if cancelledCount > 0 {
				log(fmt.Sprintf("취소 건 %d건을 제외합니다.", cancelledCount))
			}

			log("")
			log("변환을 수행하고 있습니다...")

			shippingOrders := Transform(sourceOrders)
			log(fmt.Sprintf("총 %d건의 출고 지시 데이터가 생성되었습니다.", len(shippingOrders)))

			dir := filepath.Dir(sourcePath)
			baseName := strings.TrimSuffix(filepath.Base(sourcePath), filepath.Ext(sourcePath))
			outputPath := uniqueFilename(filepath.Join(dir, baseName+"_출고지시_"+time.Now().Format("20060102")+".xlsx"))

			log("")
			log("출고 지시 엑셀 파일을 생성하고 있습니다...")

			err = WriteShippingOrders(shippingOrders, outputPath)
			if err != nil {
				log(fmt.Sprintf("[오류] 엑셀 파일 생성 실패: %v", err))
				return
			}

			log(fmt.Sprintf("[저장 위치] %s", outputPath))
			log("")

			// Run validation after transform
			log("변환 결과를 검증합니다...")
			log("")

			targetOrders, err := ReadShippingOrders(outputPath)
			if err != nil {
				log(fmt.Sprintf("[검증 오류] 출고 지시 파일 읽기 실패: %v", err))
			} else {
				result := Validate(sourceOrders, targetOrders)
				log(FormatValidation(result))
			}

			dialog.ShowConfirm("파일 열기", "변환이 완료된 파일을 바로 열어볼까요?", func(yes bool) {
				if yes {
					if err := openFile(outputPath); err != nil {
						log(fmt.Sprintf("[오류] 파일 열기 실패: %v", err))
					}
				}
			}, w)
		}, w)

		fd.SetFilter(storage.NewExtensionFileFilter([]string{".xlsx", ".xls"}))
		fd.Show()
	})

	return container.NewBorder(
		container.NewVBox(btn),
		nil, nil, nil,
		message,
	)
}

func buildValidateTab(w fyne.Window) fyne.CanvasObject {
	message := widget.NewMultiLineEntry()
	message.SetPlaceHolder("검증 결과가 여기에 표시됩니다.")
	message.Wrapping = fyne.TextWrapWord

	var sourcePath, targetPath string

	sourceLabel := widget.NewLabel("구매자 주문: (선택되지 않음)")
	targetLabel := widget.NewLabel("출고 지시: (선택되지 않음)")

	sourceBtn := widget.NewButton("구매자 주문 파일 선택", func() {
		fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil {
				message.SetText(fmt.Sprintf("[오류] %v", err))
				return
			}
			if reader == nil {
				return
			}
			reader.Close()
			sourcePath = reader.URI().Path()
			sourceLabel.SetText(fmt.Sprintf("구매자 주문: %s", filepath.Base(sourcePath)))
		}, w)
		fd.SetFilter(storage.NewExtensionFileFilter([]string{".xlsx", ".xls"}))
		fd.Show()
	})

	targetBtn := widget.NewButton("출고 지시 파일 선택", func() {
		fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil {
				message.SetText(fmt.Sprintf("[오류] %v", err))
				return
			}
			if reader == nil {
				return
			}
			reader.Close()
			targetPath = reader.URI().Path()
			targetLabel.SetText(fmt.Sprintf("출고 지시: %s", filepath.Base(targetPath)))
		}, w)
		fd.SetFilter(storage.NewExtensionFileFilter([]string{".xlsx", ".xls"}))
		fd.Show()
	})

	validateBtn := widget.NewButton("검증", func() {
		message.SetText("")
		log := func(msg string) {
			message.SetText(message.Text + msg + "\n")
		}

		if sourcePath == "" {
			log("[오류] 구매자 주문 파일을 선택해주세요.")
			return
		}
		if targetPath == "" {
			log("[오류] 출고 지시 파일을 선택해주세요.")
			return
		}

		log("파일을 읽고 있습니다...")
		log("")

		sourceOrders, err := ReadSourceOrders(sourcePath)
		if err != nil {
			log(fmt.Sprintf("[오류] 구매자 주문 파일 읽기 실패: %v", err))
			return
		}
		log(fmt.Sprintf("[구매자 주문] %s (%d건)", filepath.Base(sourcePath), len(sourceOrders)))

		targetOrders, err := ReadShippingOrders(targetPath)
		if err != nil {
			log(fmt.Sprintf("[오류] 출고 지시 파일 읽기 실패: %v", err))
			return
		}
		log(fmt.Sprintf("[출고 지시] %s (%d건)", filepath.Base(targetPath), len(targetOrders)))
		log("")

		result := Validate(sourceOrders, targetOrders)
		log(FormatValidation(result))
	})

	topPanel := container.NewVBox(
		container.NewHBox(sourceBtn, sourceLabel),
		container.NewHBox(targetBtn, targetLabel),
		validateBtn,
	)

	return container.NewBorder(
		topPanel,
		nil, nil, nil,
		message,
	)
}

func openFile(path string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", path)
	case "darwin":
		cmd = exec.Command("open", path)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", path)
	default:
		return fmt.Errorf("지원하지 않는 OS: %s", runtime.GOOS)
	}
	return cmd.Start()
}
