@echo off
title Exam Arena - Android APK Export Utility
setlocal enabledelayedexpansion

echo ======================================================================
echo                     EXAM ARENA - APK BUILD UTILITY                   
echo ======================================================================
echo.

set "PROJECT_DIR=%~dp0"
cd /d "%PROJECT_DIR%"

:: ──────────────────────────────────────────────────────────────────────────
:: 1. Auto-Detect Java / JBR
:: ──────────────────────────────────────────────────────────────────────────
echo [1/4] Checking Java Runtime Environment...

set "FOUND_JAVA="

if defined JAVA_HOME (
    if exist "%JAVA_HOME%\bin\java.exe" (
        set "FOUND_JAVA=%JAVA_HOME%"
    )
)

if "%FOUND_JAVA%"=="" (
    if exist "%LOCALAPPDATA%\Programs\Android Studio\jbr\bin\java.exe" (
        set "FOUND_JAVA=%LOCALAPPDATA%\Programs\Android Studio\jbr"
    ) else if exist "%ProgramFiles%\Android\Android Studio\jbr\bin\java.exe" (
        set "FOUND_JAVA=%ProgramFiles%\Android\Android Studio\jbr"
    ) else if exist "C:\Program Files\Android\Android Studio\jbr\bin\java.exe" (
        set "FOUND_JAVA=C:\Program Files\Android\Android Studio\jbr"
    )
)

if "%FOUND_JAVA%"=="" (
    echo.
    echo [ERROR] No Java Development Kit or Android Studio JBR found!
    echo         Please install Android Studio or JDK 21 and set JAVA_HOME.
    echo.
    pause
    exit /b 1
)

set "JAVA_HOME=%FOUND_JAVA%"
set "PATH=%JAVA_HOME%\bin;%PATH%"
echo       Found Java: %JAVA_HOME%

:: ──────────────────────────────────────────────────────────────────────────
:: 2. Auto-Detect Android SDK & Configure local.properties
:: ──────────────────────────────────────────────────────────────────────────
echo [2/4] Checking Android SDK...

set "FOUND_SDK="

if defined ANDROID_HOME (
    if exist "%ANDROID_HOME%" set "FOUND_SDK=%ANDROID_HOME%"
)
if "%FOUND_SDK%"=="" if defined ANDROID_SDK_ROOT (
    if exist "%ANDROID_SDK_ROOT%" set "FOUND_SDK=%ANDROID_SDK_ROOT%"
)
if "%FOUND_SDK%"=="" (
    if exist "%LOCALAPPDATA%\Android\Sdk" (
        set "FOUND_SDK=%LOCALAPPDATA%\Android\Sdk"
    )
)

if not "%FOUND_SDK%"=="" (
    set "ANDROID_HOME=%FOUND_SDK%"
    set "ANDROID_SDK_ROOT=%FOUND_SDK%"
    echo       Found Android SDK: %FOUND_SDK%
    
    :: Write local.properties with escaped backslashes
    set "SDK_ESCAPED=%FOUND_SDK:\=\\%"
    echo sdk.dir=!SDK_ESCAPED!> "%PROJECT_DIR%\local.properties"
) else (
    echo [WARN] Android SDK location not detected automatically.
    echo        Ensure sdk.dir is defined in %PROJECT_DIR%\local.properties
)

:: ──────────────────────────────────────────────────────────────────────────
:: 3. Ensure Gradle Wrapper Jar
:: ──────────────────────────────────────────────────────────────────────────
echo [3/4] Verifying Gradle wrapper...

if not exist "%PROJECT_DIR%\gradle\wrapper\gradle-wrapper.jar" (
    echo       Downloading official gradle-wrapper.jar...
    powershell -NoProfile -Command "Invoke-WebRequest -Uri 'https://raw.githubusercontent.com/gradle/gradle/v8.11.1/gradle/wrapper/gradle-wrapper.jar' -OutFile '%PROJECT_DIR%\gradle\wrapper\gradle-wrapper.jar'"
)

:: ──────────────────────────────────────────────────────────────────────────
:: 4. Build APK
:: ──────────────────────────────────────────────────────────────────────────
echo [4/4] Building Debug APK (assembleDebug)...
echo       This may take a minute on the first run while Gradle syncs...
echo.

call "%PROJECT_DIR%\gradlew.bat" assembleDebug --no-daemon

if %ERRORLEVEL% NEQ 0 (
    echo.
    echo ======================================================================
    echo [BUILD FAILED] Gradle encountered an error building the APK.
    echo ======================================================================
    echo Please check the error log above.
    echo Tip: Ensure Android SDK platforms and build-tools are installed via
    echo      Android Studio -> SDK Manager.
    echo.
    pause
    exit /b %ERRORLEVEL%
)

:: ──────────────────────────────────────────────────────────────────────────
:: 5. Locate & Export Output APK
:: ──────────────────────────────────────────────────────────────────────────
set "OUTPUT_DIR=%PROJECT_DIR%\build_output"
if not exist "%OUTPUT_DIR%" mkdir "%OUTPUT_DIR%"

set "SRC_APK=%PROJECT_DIR%\app\build\outputs\apk\debug\app-debug.apk"
set "DEST_APK=%OUTPUT_DIR%\ExamArena-debug.apk"

if exist "%SRC_APK%" (
    copy /y "%SRC_APK%" "%DEST_APK%" > nul
    echo.
    echo ======================================================================
    echo                      APK EXPORTED SUCCESSFULLY!                       
    echo ======================================================================
    echo.
    echo Exported APK Location:
    echo   %DEST_APK%
    echo.
    
    :: Open build_output folder
    explorer.exe "%OUTPUT_DIR%"
) else (
    echo [WARN] Built successfully, but APK was not found at standard path:
    echo        %SRC_APK%
)

echo.
echo You can install this APK on any Android phone or emulator with:
echo   adb install -r "%DEST_APK%"
echo.
pause
