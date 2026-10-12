@echo off
REM build-windows.bat
REM 双击或在 cmd 中执行即可。
REM 生成产物: build\bin\simple2tradition-amd64-installer.exe

setlocal enabledelayedexpansion
cd /d "%~dp0\.."

echo.
echo ==^> 检查 wails CLI
where wails >nul 2>&1
if errorlevel 1 (
    echo [错误] 未找到 wails，请先: go install github.com/wailsapp/wails/v2/cmd/wails@latest
    exit /b 1
)

echo.
echo ==^> 检查 NSIS (makensis)
where makensis >nul 2>&1
if errorlevel 1 (
    echo [错误] 未找到 makensis，请先安装 NSIS 并把 Bin 目录加入 PATH
    exit /b 1
)

echo.
echo ==^> 清理 build\bin
if exist build\bin rmdir /s /q build\bin

echo.
echo ==^> 构建前端
cd frontend
if not exist node_modules (
    call npm install || goto :error
)
call npm run build || goto :error
cd ..

echo.
echo ==^> 构建 Windows 安装包
wails build -nsis -platform windows/amd64 -clean || goto :error

echo.
echo ==^> 完成
if exist build\bin\simple2tradition-amd64-installer.exe (
    for %%I in (build\bin\simple2tradition-amd64-installer.exe) do echo 安装包: build\bin\%%~nxI (%%~zI 字节)
) else (
    echo [警告] 未在预期位置找到安装包，请检查 build\bin 目录。
)

endlocal
exit /b 0

:error
echo.
echo [失败] 构建过程中断，错误码 %errorlevel%
endlocal
exit /b 1
