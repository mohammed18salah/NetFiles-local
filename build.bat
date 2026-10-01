@echo off
REM NetFiles Build Script — Created by Mohammed Salah
REM Produces a static netfiles.exe with no runtime dependencies

echo.
echo  === Building NetFiles ===
echo  Created by Mohammed Salah
echo.

REM Ensure we're in the project root
cd /d "%~dp0"

REM Tidy dependencies
echo [1/3] Resolving dependencies...
go mod tidy
if %ERRORLEVEL% neq 0 (
    echo [FAIL] Failed to resolve dependencies
    exit /b 1
)

REM Build with stripped symbols for smallest binary
echo [2/3] Compiling...
go build -ldflags "-s -w" -o netfiles.exe .
if %ERRORLEVEL% neq 0 (
    echo [FAIL] Build failed
    exit /b 1
)

REM Show result
echo [3/3] Done!
echo.
for %%A in (netfiles.exe) do echo  Output: netfiles.exe (%%~zA bytes)
echo.
