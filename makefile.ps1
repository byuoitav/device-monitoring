$COMMAND = $args[0]

# =============================
# Metadata
# =============================
$NAME = "device-monitoring"
$OWNER = "byuoitav"
$BUILD_DIR = "dist"
$BIN_OUTPUT = "$BUILD_DIR/$NAME"
$NG1 = "dashboard"

$BRANCH = Invoke-Expression "git rev-parse --abbrev-ref HEAD"
Write-Output "Branch: $BRANCH"

$COMMIT_HASH = Invoke-Expression "git rev-parse --short HEAD"
$TAG = $COMMIT_HASH
try {
    $NEW_TAG = (git describe --exact-match --tags HEAD 2>$null)
    if ($NEW_TAG.Length -gt 0) {
        $TAG = $NEW_TAG
        Write-Host "Tag: $TAG"
    }
}
catch {
    Write-Host "No exact tag on HEAD, using commit hash: $TAG"
}

# =============================
# Optional Go build flag overrides (set as env vars to override defaults)
# =============================
$GOOS = "linux"
$GOARCH = "arm"
$GOARM = "7"
$CGO_ENABLED = "0"
$TAGS = ""
$LDFLAGS = ""
$GCFLAGS = ""
$ASMFLAGS = ""
$BUILD_FLAGS = ""
$MAIN_PKG = "."

# Compose optional build flags
$COMMON_BUILD_FLAGS = ""
if ($TAGS) { $COMMON_BUILD_FLAGS += " -tags '$TAGS'" }
if ($LDFLAGS) { $COMMON_BUILD_FLAGS += " -ldflags '$LDFLAGS'" }
if ($GCFLAGS) { $COMMON_BUILD_FLAGS += " -gcflags '$GCFLAGS'" }
if ($ASMFLAGS) { $COMMON_BUILD_FLAGS += " -asmflags '$ASMFLAGS'" }
if ($BUILD_FLAGS) { $COMMON_BUILD_FLAGS += " $BUILD_FLAGS" }

# =============================
# Functions
# =============================

function All {
    Write-Output "Running all: build-web then build-local"
    Build-Web
    Build-Local
}

function Build-Local {
    Write-Output "Building $NAME for linux/arm (GOARM=7) - override with GOOS/GOARCH/GOARM/CGO_ENABLED env vars as needed..."

    if (-not (Test-Path -Path $BUILD_DIR)) {
        New-Item -Path $BUILD_DIR -ItemType Directory | Out-Null
    }

    Set-Item -Path env:CGO_ENABLED -Value $CGO_ENABLED
    Set-Item -Path env:GOOS        -Value $GOOS
    Set-Item -Path env:GOARCH      -Value $GOARCH
    Set-Item -Path env:GOARM       -Value $GOARM

    Invoke-Expression "go build $COMMON_BUILD_FLAGS -o $BIN_OUTPUT -v $MAIN_PKG"
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Build failed"
        exit 1
    }

    Write-Output "Build complete: $BIN_OUTPUT"

    # Restore env to Windows defaults
    Set-Item -Path env:GOOS        -Value "windows"
    Set-Item -Path env:GOARCH      -Value "amd64"
    Remove-Item -Path env:GOARM    -ErrorAction SilentlyContinue
}

function Build-Binaries {
    $PLATFORMS = @("linux/amd64", "linux/arm")
    Write-Output "Building binaries for: $($PLATFORMS -join ', ')"

    if (-not (Test-Path -Path $BUILD_DIR)) {
        New-Item -Path $BUILD_DIR -ItemType Directory | Out-Null
    }

    foreach ($platform in $PLATFORMS) {
        $parts = $platform -split "/"
        $OS = $parts[0]
        $ARCH = $parts[1]
        $OUT = "$BUILD_DIR/$NAME-$OS-$ARCH"

        Write-Output "*****************************************"
        Write-Output "Building for $OS/$ARCH => $OUT"

        Set-Item -Path env:CGO_ENABLED -Value $CGO_ENABLED
        Set-Item -Path env:GOOS        -Value $OS
        Set-Item -Path env:GOARCH      -Value $ARCH

        if ($ARCH -eq "arm") {
            Set-Item -Path env:GOARM -Value $GOARM
        }
        else {
            Remove-Item -Path env:GOARM -ErrorAction SilentlyContinue
        }

        Invoke-Expression "go build $COMMON_BUILD_FLAGS -o $OUT -v $MAIN_PKG"
        if ($LASTEXITCODE -ne 0) {
            Write-Error "Build failed for $OS/$ARCH"
            exit 1
        }
    }

    Write-Output "All binaries built in ./$BUILD_DIR/"

    # Restore env to Windows defaults
    Set-Item -Path env:GOOS  -Value "windows"
    Set-Item -Path env:GOARCH -Value "amd64"
    Remove-Item -Path env:GOARM -ErrorAction SilentlyContinue
}

function Build-Web {
    Write-Output "Preparing static dashboard assets..."

    $dest = "files/$NG1"
    if (-not (Test-Path -Path $dest)) {
        New-Item -Path $dest -ItemType Directory -Force | Out-Null
    }

    # Copy dashboard/ -> files/dashboard/, excluding node_modules and dist
    Copy-Item -Path "dashboard/*" -Destination $dest -Recurse -Force

    # Remove excluded directories from destination
    $excludeDirs = @("node_modules", "dist")
    foreach ($dir in $excludeDirs) {
        $excludePath = Join-Path $dest $dir
        if (Test-Path -Path $excludePath) {
            Remove-Item -Path $excludePath -Recurse -Force
            Write-Output "  Removed $excludePath"
        }
    }

    Write-Output "Dashboard assets ready in $dest"
}

function Clean {
    Write-Output "Cleaning build artifacts..."

    $toRemove = @($BUILD_DIR, "files", "vendor")
    foreach ($item in $toRemove) {
        if (Test-Path -Path $item) {
            Remove-Item -Path $item -Recurse -Force
            Write-Output "  Removed $item/"
        }
    }

    Get-ChildItem -Filter "*.tar.gz" | Remove-Item -Force
    Write-Output "Clean complete."
}

function Deps {
    Write-Output "Downloading Go dependencies..."
    Invoke-Expression "go mod download"
}

function Package {
    param([string]$TagName)

    $tarball = "dmm-$TagName.tar.gz"
    Write-Host "Packaging $tarball"

    # Always build fresh before packaging
    Build-Local
    Build-Web

    # Copy supporting files into the files/ directory
    Copy-Item -Path "version.txt"        -Destination "files/" -Force
    Copy-Item -Path "service-config.json" -Destination "files/" -Force

    # Write version file for the dashboard to display
    Set-Content -Path "files/$NG1/version.json" -Value "{`"version`": `"$TagName`"}"

    # Create tarball (requires tar, available in Windows 10 1803+)
    $tarArgs = "-czf `"$tarball`" -C `"$(Get-Location)`" files -C `"$BUILD_DIR`" $NAME"
    Invoke-Expression "tar $tarArgs"
    if ($LASTEXITCODE -ne 0) {
        Write-Error "tar failed"
        exit 1
    }

    return $tarball
}

function Deploy-Local {
    $tarball = Package -TagName $TAG
    Write-Output "Tarball created locally: $tarball"
}

function Deploy {
    $tarball = Package -TagName $TAG

    # Upload as a GitHub release asset using the gh CLI
    Write-Output "Uploading $tarball to GitHub release '$TAG'..."
    $releaseExists = gh release view $TAG --repo "$OWNER/$NAME" 2>&1
    if ($LASTEXITCODE -ne 0) {
        Write-Output "Release '$TAG' not found - creating it..."
        gh release create $TAG $tarball --repo "$OWNER/$NAME" --title $TAG --notes ""
    }
    else {
        gh release upload $TAG $tarball --repo "$OWNER/$NAME" --clobber
    }
    if ($LASTEXITCODE -ne 0) {
        Write-Error "GitHub release upload failed"
        exit 1
    }

    Write-Output "Deploy complete."
}

# =============================
# Command Dispatch
# =============================
switch ($COMMAND) {
    "all" { All }
    "build-local" { Build-Local }
    "build-binaries" { Build-Binaries }
    "build-web" { Build-Web }
    "clean" { Clean }
    "deps" { Deps }
    "deploy-local" { Deploy-Local }
    "deploy" { Deploy }
    default {
        Write-Output "Usage: .\makefile.ps1 <target>"
        Write-Output ""
        Write-Output "Targets:"
        Write-Output "  all             Build web assets then binary (linux/arm default)"
        Write-Output "  build-local     Build binary for linux/arm (GOARM=7) - override via env vars"
        Write-Output "  build-binaries  Build binaries for linux/amd64 and linux/arm"
        Write-Output "  build-web       Sync dashboard/ -> files/dashboard/"
        Write-Output "  clean           Remove dist/, files/, vendor/, *.tar.gz"
        Write-Output "  deps            Download Go module dependencies"
        Write-Output "  deploy-local    Package into a local tar.gz without uploading"
        Write-Output "  deploy          Package and upload as a GitHub release (requires gh CLI)"
    }
}
