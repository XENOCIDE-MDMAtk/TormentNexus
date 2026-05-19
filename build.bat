@echo off
setlocal

echo Building Borg AIOS...

REM Build Go sidecar
where go >nul 2>nul
if errorlevel 1 (
    echo [SKIP] Go not found
    goto :ts_build
)

echo [1/2] Building Go sidecar...
for /f "tokens=*" %%v in ('type VERSION') do set VER=%%v
cd go
go build -ldflags "-s -w -X github.com/borghq/borg-go/internal/buildinfo.Version=%VER%" -buildvcs=false -o ..\bin\borg.exe ./cmd/borg
if errorlevel 1 (
    echo [FAIL] Go build failed
    cd ..
    exit /b 1
)
cd ..
echo OK - bin\borg.exe built

:ts_build
REM Build TypeScript core
where npx >nul 2>nul
if errorlevel 1 (
    echo [SKIP] npx not found
    goto :done
)

echo [2/2] Building TypeScript core...
cd packages\core
call npx tsc --project tsconfig.json
if errorlevel 1 (
    echo [FAIL] TypeScript build failed
    cd ..\..
    exit /b 1
)
cd ..\..
echo OK - packages\core built

:done
echo.
echo Build complete.
