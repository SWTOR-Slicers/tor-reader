@REM fyne package -os windows
@REM fyne package -os darwin
@REM fyne package -os linux

garble build -o tor-reader-gui.exe -ldflags "-s -w -H=windowsgui"