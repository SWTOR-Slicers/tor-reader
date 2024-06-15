package main

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	tor_reader "github.com/SWTOR-Slicers/tor-reader"
)

// FileOpen

func main() {

	outputFolder := "./output"
	hashFile := ""

	a := app.New()
	w := a.NewWindow("Tor Extractor")
	w.Resize(fyne.NewSize(800, 600))

	data := binding.BindStringList(
		&[]string{},
	)

	list := widget.NewListWithData(data,
		func() fyne.CanvasObject {
			return widget.NewLabel("template")
		},
		func(i binding.DataItem, o fyne.CanvasObject) {
			o.(*widget.Label).Bind(i.(binding.String))
		})

	add := widget.NewButton("Add Tor", func() {
		dialog.ShowFileOpen(func(uri fyne.URIReadCloser, err error) {
			if err != nil {
				return
			}

			if uri == nil {
				return
			}

			path := uri.URI().String()
			fmt.Println(path)

			path = strings.Replace(path, "file://", "", 1)
			// torReader := tor_reader.NewTorReader(uri.URI().String())
			// torReader := tor_reader.NewTorReader(path)
			// torReader.Read()

			data.Append(path)
		}, w)

	})

	extract := widget.NewButton("Extract", func() {
		fmt.Println("extract")

		files, err := data.Get()

		if err != nil {
			dialog.ShowError(err, w)
			return
		}

		if len(files) == 0 {
			dialog.ShowError(errors.New("no files selected"), w)
			return
		}

		if hashFile == "" {
			dialog.ShowError(errors.New("no hash file selected"), w)
			return
		}

		results := tor_reader.StartExtractor(files, outputFolder, hashFile, runtime.NumCPU())

		resultString := fmt.Sprintf("Files Attempted: %d\nFiles No Hash: %d\nTime Taken: %s", results.FilesAttempted, results.FilesNoHash, results.TimeTaken)

		dialog.ShowInformation("Extraction Complete", resultString, w)
	})

	setHash := widget.NewButton("Set Hash", func() {
		fmt.Println("set output")

		fileOpener := dialog.NewFileOpen(func(uri fyne.URIReadCloser, err error) {
			if err != nil {
				return
			}

			if uri == nil {
				return
			}

			path := uri.URI().String()
			fmt.Println(path)

			path = strings.Replace(path, "file://", "", 1)
			hashFile = path
		}, w)

		fileOpener.SetFilter(storage.NewExtensionFileFilter([]string{".txt"}))

		// get the current working directory
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Println(err)
		}

		actualFolderURI := storage.NewFileURI(cwd)
		parentLister, err := storage.ListerForURI(actualFolderURI)
		if err == nil {
			fileOpener.SetLocation(parentLister)
		}

		fileOpener.Show()
	})

	setOuptut := widget.NewButton("Set Output", func() {

		// newWindow := a.NewWindow("Output")
		// newWindow.Resize(fyne.NewSize(800, 600))

		dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
			// newWindow.Close()
			if err != nil {
				return
			}

			if uri == nil {
				return
			}

			path := uri.String()
			fmt.Println(path)
			// torReader := tor_reader.NewTorReader(uri.URI().String())
			// torReader := tor_reader.NewTorReader(path)
			// torReader.Read()

			// remove the file:// from the path
			path = strings.Replace(path, "file://", "", 1)

			outputFolder = path
		}, w)

		// newWindow.Show()
	})

	w.SetContent(container.NewBorder(nil, container.NewGridWithRows(4, add, setOuptut, setHash, extract), nil, nil, list))
	w.Resize(fyne.NewSize(800, 600))

	w.SetOnDropped(func(pos fyne.Position, uris []fyne.URI) {
		for _, uri := range uris {
			fmt.Println(uri.String())

			path := uri.String()

			// remove the file:// from the path
			path = strings.Replace(path, "file://", "", 1)

			data.Append(path)
		}
	})

	w.SetMaster()

	w.ShowAndRun()
}
