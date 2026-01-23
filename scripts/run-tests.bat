@echo off
setlocal enabledelayedexpansion

REM Change to project root directory
cd /d "%~dp0.."

echo ===================================
echo AI Commit Hub - Test Runner
echo ===================================

set RUN_UNIT=1
set RUN_INTEGRATION=1
set RUN_FRONTEND=0
set COVERAGE=0

:parse_args
if "%~1"=="--no-unit" set RUN_UNIT=0
if "%~1"=="--no-integration" set RUN_INTEGRATION=0
if "%~1"=="--frontend" set RUN_FRONTEND=1
if "%~1"=="--coverage" set COVERAGE=1
shift
if not "%~1"=="" goto parse_args

if not exist "tmp\test-results" mkdir tmp\test-results

set TOTAL_TESTS=0
set PASSED_TESTS=0
set FAILED_TESTS=0

if %RUN_UNIT%==1 (
    echo.
    echo [1/2] Running Backend Unit Tests...
    echo -----------------------------------

    go test ./pkg/git/... -v > tmp\test-results\unit-git.log 2>&1
    if !errorlevel!==0 (
        echo [PASS] Git tests
        set /a PASSED_TESTS+=1
    ) else (
        echo [FAIL] Git tests - see tmp\test-results\unit-git.log
        set /a FAILED_TESTS+=1
    )
    set /a TOTAL_TESTS+=1

    go test ./pkg/service/... -v > tmp\test-results\unit-service.log 2>&1
    if !errorlevel!==0 (
        echo [PASS] Service tests
        set /a PASSED_TESTS+=1
    ) else (
        echo [FAIL] Service tests - see tmp\test-results\unit-service.log
        set /a FAILED_TESTS+=1
    )
    set /a TOTAL_TESTS+=1
)

if %RUN_INTEGRATION%==1 (
    echo.
    echo [2/2] Running Integration Tests...
    echo -----------------------------------

    go test ./tests/integration/... -v > tmp\test-results\integration.log 2>&1
    if !errorlevel!==0 (
        echo [PASS] Integration tests
        set /a PASSED_TESTS+=1
    ) else (
        echo [FAIL] Integration tests - see tmp\test-results\integration.log
        set /a FAILED_TESTS+=1
    )
    set /a TOTAL_TESTS+=1
)

echo.
echo ===================================
echo Test Summary
echo ===================================
echo Total Suites: %TOTAL_TESTS%
echo Passed: %PASSED_TESTS%
echo Failed: %FAILED_TESTS%
echo ===================================

if %FAILED_TESTS%==0 (
    echo.
    echo All tests passed!
    exit /b 0
) else (
    echo.
    echo Some tests failed. Check logs in tmp\test-results\
    exit /b 1
)
