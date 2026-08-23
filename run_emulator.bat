@echo off
title Exam Arena - Android Emulator Runner
setlocal enabledelayedexpansion

echo ======================================================================
echo                EXAM ARENA - ANDROID EMULATOR RUNNER                  
echo ======================================================================
echo.

:: ──────────────────────────────────────────────────────────────────────────
:: 0. Determine Project and Root Directories
:: ──────────────────────────────────────────────────────────────────────────
set "SCRIPT_DIR=%~dp0"
if exist "%SCRIPT_DIR%frontend_app\build.gradle.kts" (
    set "PROJECT_DIR=%SCRIPT_DIR%frontend_app"
    set "ROOT_DIR=%SCRIPT_DIR%"
) else if exist "%SCRIPT_DIR%app\build.gradle.kts" (
    set "PROJECT_DIR=%SCRIPT_DIR%"
    set "ROOT_DIR=%SCRIPT_DIR%..\"
) else (
    set "PROJECT_DIR=%SCRIPT_DIR%frontend_app"
    set "ROOT_DIR=%SCRIPT_DIR%"
)

cd /d "%PROJECT_DIR%"

:: ──────────────────────────────────────────────────────────────────────────
:: 1. Auto-Detect Android SDK & Emulator tools (adb, emulator, avdmanager)
:: ──────────────────────────────────────────────────────────────────────────
echo [1/6] Detecting Android SDK and Emulator tools...

set "FOUND_SDK="

if defined ANDROID_HOME (
    if exist "%ANDROID_HOME%" set "FOUND_SDK=%ANDROID_HOME%"
)
if "%FOUND_SDK%"=="" if defined ANDROID_SDK_ROOT (
    if exist "%ANDROID_SDK_ROOT%" set "FOUND_SDK=%ANDROID_SDK_ROOT%"
)

:: Try reading sdk.dir from local.properties if available
if "%FOUND_SDK%"=="" if exist "%PROJECT_DIR%\local.properties" (
    for /f "tokens=1,* delims==" %%A in ('type "%PROJECT_DIR%\local.properties" 2^>nul') do (
        if "%%A"=="sdk.dir" (
            set "PROP_SDK=%%B"
            set "PROP_SDK=!PROP_SDK:\\=\!"
            if exist "!PROP_SDK!" set "FOUND_SDK=!PROP_SDK!"
        )
    )
)

:: Common Android SDK installation paths on Windows
if "%FOUND_SDK%"=="" (
    if exist "D:\Android\Sdk" (
        set "FOUND_SDK=D:\Android\Sdk"
    ) else if exist "%LOCALAPPDATA%\Android\Sdk" (
        set "FOUND_SDK=%LOCALAPPDATA%\Android\Sdk"
    ) else if exist "C:\Android\Sdk" (
        set "FOUND_SDK=C:\Android\Sdk"
    ) else if exist "E:\Android\Sdk" (
        set "FOUND_SDK=E:\Android\Sdk"
    ) else if exist "%ProgramFiles%\Android\Android Studio\sdk" (
        set "FOUND_SDK=%ProgramFiles%\Android\Android Studio\sdk"
    )
)

if "%FOUND_SDK%"=="" (
    echo.
    echo [ERROR] Android SDK not found!
    echo         Please ensure Android Studio / Android SDK is installed.
    echo.
    pause
    exit /b 1
)

set "ANDROID_HOME=%FOUND_SDK%"
set "ANDROID_SDK_ROOT=%FOUND_SDK%"
set "PATH=%ANDROID_HOME%\emulator;%ANDROID_HOME%\platform-tools;%ANDROID_HOME%\cmdline-tools\latest\bin;%PATH%"

set "ADB_EXE=%ANDROID_HOME%\platform-tools\adb.exe"
set "EMULATOR_EXE=%ANDROID_HOME%\emulator\emulator.exe"
set "AVDMANAGER_EXE=%ANDROID_HOME%\cmdline-tools\latest\bin\avdmanager.bat"

if not exist "%ADB_EXE%" (
    echo [ERROR] adb.exe not found at %ADB_EXE%!
    pause
    exit /b 1
)

echo       SDK Directory: %ANDROID_HOME%
echo       ADB Location:  %ADB_EXE%

:: ──────────────────────────────────────────────────────────────────────────
:: 2. Auto-Detect Java & Configure local.properties
:: ──────────────────────────────────────────────────────────────────────────
echo [2/6] Detecting Java Runtime Environment...

set "FOUND_JAVA="
if defined JAVA_HOME (
    if exist "%JAVA_HOME%\bin\java.exe" set "FOUND_JAVA=%JAVA_HOME%"
)
if "%FOUND_JAVA%"=="" (
    if exist "%LOCALAPPDATA%\Programs\Android Studio\jbr\bin\java.exe" (
        set "FOUND_JAVA=%LOCALAPPDATA%\Programs\Android Studio\jbr"
    ) else if exist "%ProgramFiles%\Android\Android Studio\jbr\bin\java.exe" (
        set "FOUND_JAVA=%ProgramFiles%\Android\Android Studio\jbr"
    ) else if exist "C:\Program Files\Android\Android Studio\jbr\bin\java.exe" (
        set "FOUND_JAVA=C:\Program Files\Android\Android Studio\jbr"
    ) else if exist "D:\Android\Android Studio\jbr\bin\java.exe" (
        set "FOUND_JAVA=D:\Android\Android Studio\jbr"
    )
)

:: Fallback check system PATH for java.exe
if "%FOUND_JAVA%"=="" (
    for /f "tokens=*" %%J in ('where java 2^>nul') do (
        if "%FOUND_JAVA%"=="" (
            for %%P in ("%%~dpJ..") do set "FOUND_JAVA=%%~fP"
        )
    )
)

if not "%FOUND_JAVA%"=="" (
    set "JAVA_HOME=%FOUND_JAVA%"
    set "PATH=%JAVA_HOME%\bin;%PATH%"
    echo       Java Location: %JAVA_HOME%
)

set "SDK_ESCAPED=%ANDROID_HOME:\=\\%"
echo sdk.dir=!SDK_ESCAPED!> "%PROJECT_DIR%\local.properties"

:: ──────────────────────────────────────────────────────────────────────────
:: 3. Check / Start Backend Service
:: ──────────────────────────────────────────────────────────────────────────
echo [3/6] Checking Backend Server...

powershell -NoProfile -Command "try { $r = Invoke-WebRequest -Uri 'http://localhost:8080/health' -TimeoutSec 2 -UseBasicParsing; if ($r.StatusCode -eq 200) { exit 0 } else { exit 1 } } catch { exit 1 }" >nul 2>&1
if %ERRORLEVEL% EQU 0 (
    echo       Backend is already running on http://localhost:8080 [OK]
) else (
    echo       Backend is not running. Starting Go Backend server in new window...
    if exist "%ROOT_DIR%backend" (
        start "Exam Arena - Backend Server" cmd /k "cd /d "%ROOT_DIR%backend" && go run ./cmd/server"
        echo       Waiting 3 seconds for backend initialization...
        timeout /t 3 /nobreak >nul
    ) else (
        echo       [WARN] Backend folder not found at %ROOT_DIR%backend. Please start it manually.
    )
)

:: ──────────────────────────────────────────────────────────────────────────
:: 4. Detect / Launch Android Emulator
:: ──────────────────────────────────────────────────────────────────────────
echo [4/6] Checking Android Devices and Emulators...

"%ADB_EXE%" start-server >nul 2>&1

:: Check if an emulator/device is already connected
set "DEVICE_READY=0"
set "CONNECTED_DEVICE="
for /f "skip=1 tokens=1,2" %%A in ('"%ADB_EXE%" devices') do (
    if "%%B"=="device" (
        set "DEVICE_READY=1"
        set "CONNECTED_DEVICE=%%A"
    )
)

if "!DEVICE_READY!"=="1" (
    echo       Found active running device/emulator: !CONNECTED_DEVICE!
) else (
    echo       No active device detected. Looking for available AVDs...
    
    set "TARGET_AVD="
    if exist "%EMULATOR_EXE%" (
        for /f "tokens=*" %%V in ('"%EMULATOR_EXE%" -list-avds 2^>nul') do (
            if "!TARGET_AVD!"=="" set "TARGET_AVD=%%V"
        )
    )

    :: Auto-create ExamArenaAVD if no AVD is installed
    if "!TARGET_AVD!"=="" (
        if exist "%AVDMANAGER_EXE%" (
            echo       No AVD found. Auto-creating ExamArenaAVD...
            echo no | "%AVDMANAGER_EXE%" create avd -n ExamArenaAVD -k "system-images;android-37.0;google_apis_playstore_ps16k;arm64-v8a" --force >nul 2>&1
            for /f "tokens=*" %%V in ('"%EMULATOR_EXE%" -list-avds 2^>nul') do (
                if "!TARGET_AVD!"=="" set "TARGET_AVD=%%V"
            )
        )
    )

    if not "!TARGET_AVD!"=="" (
        echo       Launching AVD: !TARGET_AVD!...
        start "Android Emulator" "%EMULATOR_EXE%" -avd "!TARGET_AVD!" -netdelay none -netspeed full
        echo       Waiting for emulator to connect via adb...
        
        "%ADB_EXE%" wait-for-device
        
        echo       Waiting for Android OS boot completion...
        set /a BOOT_RETRY=0
        :BOOT_CHECK_LOOP
        set "BOOT_PROP=0"
        for /f "tokens=1 delims= " %%B in ('"%ADB_EXE%" shell getprop sys.boot_completed 2^>nul') do (
            set "BOOT_RAW=%%B"
            set "BOOT_PROP=!BOOT_RAW:~0,1!"
        )
        if not "!BOOT_PROP!"=="1" (
            set /a BOOT_RETRY+=1
            if !BOOT_RETRY! GTR 60 (
                echo       [WARN] Boot check timed out. Attempting to proceed...
                goto BOOT_DONE
            )
            timeout /t 2 /nobreak >nul
            goto BOOT_CHECK_LOOP
        )
        echo       Emulator booted successfully! [OK]
        :BOOT_DONE
    ) else (
        echo.
        echo [WARN] No AVD (Android Virtual Device) found!
        echo        Please create an AVD in Android Studio -> Device Manager,
        echo        or connect a physical Android device with USB Debugging enabled.
        echo.
        echo Press any key once your device/emulator is connected...
        pause >nul
    )
)

:: ──────────────────────────────────────────────────────────────────────────
:: 5. Build and Install APK
:: ──────────────────────────────────────────────────────────────────────────
echo [5/6] Building & Installing Exam Arena App...

if not exist "%PROJECT_DIR%\gradle.properties" (
    echo       Creating gradle.properties with AndroidX support...
    echo android.useAndroidX=true> "%PROJECT_DIR%\gradle.properties"
    echo android.nonTransitiveRClass=true>> "%PROJECT_DIR%\gradle.properties"
)

if not exist "%PROJECT_DIR%\gradle\wrapper\gradle-wrapper.jar" (
    echo       Downloading official gradle-wrapper.jar...
    powershell -NoProfile -Command "Invoke-WebRequest -Uri 'https://raw.githubusercontent.com/gradle/gradle/v8.11.1/gradle/wrapper/gradle-wrapper.jar' -OutFile '%PROJECT_DIR%\gradle\wrapper\gradle-wrapper.jar'"
)

call "%PROJECT_DIR%\gradlew.bat" assembleDebug --no-daemon

if %ERRORLEVEL% NEQ 0 (
    echo [ERROR] Gradle compilation failed. Please review output above.
    pause
    exit /b %ERRORLEVEL%
)

set "APK_FILE=%PROJECT_DIR%\app\build\outputs\apk\debug\app-debug.apk"
if not exist "%APK_FILE%" (
    set "APK_FILE=%PROJECT_DIR%\build_output\ExamArena-debug.apk"
)

if not exist "%APK_FILE%" (
    echo [ERROR] APK not found at %APK_FILE%!
    pause
    exit /b 1
)

echo       Installing APK on device...
"%ADB_EXE%" install -r -d "%APK_FILE%"

if %ERRORLEVEL% NEQ 0 (
    echo [ERROR] Failed to install APK via adb!
    pause
    exit /b %ERRORLEVEL%
)
echo       App installed successfully! [OK]

:: ──────────────────────────────────────────────────────────────────────────
:: 6. Launch App & Setup Port Forwarding
:: ──────────────────────────────────────────────────────────────────────────
echo [6/6] Launching Exam Arena on Emulator...

:: Forward port 8080 so both 10.0.2.2 and localhost:8080 work identically
"%ADB_EXE%" reverse tcp:8080 tcp:8080 >nul 2>&1

:: Start the main activity
"%ADB_EXE%" shell am start -n com.examarena.debug/com.examarena.MainActivity >nul 2>&1

if %ERRORLEVEL% NEQ 0 (
    "%ADB_EXE%" shell am start -n com.examarena/com.examarena.MainActivity >nul 2>&1
)

if %ERRORLEVEL% NEQ 0 (
    "%ADB_EXE%" shell monkey -p com.examarena.debug -c android.intent.category.LAUNCHER 1 >nul 2>&1
)

echo.
echo ======================================================================
echo                  EXAM ARENA IS RUNNING IN EMULATOR!                   
echo ======================================================================
echo.
echo  App Package:      com.examarena.debug (or com.examarena)
echo  Backend Target:   http://10.0.2.2:8080 (or http://localhost:8080)
echo  Realtime WS:      ws://10.0.2.2:8080/ws
echo.
echo  Live Logcat Output (Ctrl+C to stop):
echo ======================================================================

"%ADB_EXE%" logcat -s "ExamArena" "ExamArenaWS" AndroidRuntime:E