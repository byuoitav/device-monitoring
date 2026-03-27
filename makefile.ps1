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
    Write-Host "Running all: build-web then build-local"
    Build-Web
    Build-Local
}

function Build-Local {
    Write-Host "Building $NAME for linux/arm (GOARM=7) - override with GOOS/GOARCH/GOARM/CGO_ENABLED env vars as needed..."

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

    Write-Host "Build complete: $BIN_OUTPUT"

    # Restore env to Windows defaults
    Set-Item -Path env:GOOS        -Value "windows"
    Set-Item -Path env:GOARCH      -Value "amd64"
    Remove-Item -Path env:GOARM    -ErrorAction SilentlyContinue
}

function Build-Binaries {
    $PLATFORMS = @("linux/amd64", "linux/arm")
    Write-Host "Building binaries for: $($PLATFORMS -join ', ')"

    if (-not (Test-Path -Path $BUILD_DIR)) {
        New-Item -Path $BUILD_DIR -ItemType Directory | Out-Null
    }

    foreach ($platform in $PLATFORMS) {
        $parts = $platform -split "/"
        $OS = $parts[0]
        $ARCH = $parts[1]
        $OUT = "$BUILD_DIR/$NAME-$OS-$ARCH"

        Write-Host "*****************************************"
        Write-Host "Building for $OS/$ARCH => $OUT"

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

    Write-Host "All binaries built in ./$BUILD_DIR/"

    # Restore env to Windows defaults
    Set-Item -Path env:GOOS  -Value "windows"
    Set-Item -Path env:GOARCH -Value "amd64"
    Remove-Item -Path env:GOARM -ErrorAction SilentlyContinue
}

function Build-Web {
    Write-Host "Preparing static dashboard assets..."

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
            Write-Host "  Removed $excludePath"
        }
    }

    Write-Host "Dashboard assets ready in $dest"
}

function Clean {
    Write-Host "Cleaning build artifacts..."

    $toRemove = @($BUILD_DIR, "files", "vendor")
    foreach ($item in $toRemove) {
        if (Test-Path -Path $item) {
            Remove-Item -Path $item -Recurse -Force
            Write-Host "  Removed $item/"
        }
    }

    Get-ChildItem -Filter "*.tar.gz" | Remove-Item -Force
    Write-Host "Clean complete."
}

function Deps {
    Write-Host "Downloading Go dependencies..."
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
    Write-Host "Tarball created locally: $tarball"
}

function Deploy {
    $tarball = Package -TagName $TAG

    # Require a GitHub token
    $token = $env:GITHUB_TOKEN
    if (-not $token) {
        Write-Error "GITHUB_TOKEN environment variable is not set. Export it before running deploy."
        exit 1
    }

    $headers = @{
        Authorization          = "Bearer $token"
        Accept                 = "application/vnd.github+json"
        "X-GitHub-Api-Version" = "2022-11-28"
    }

    # Check whether the release already exists
    Write-Host "Uploading $tarball to GitHub release '$TAG'..."
    $releaseId = $null
    try {
        $release = Invoke-RestMethod -Uri "https://api.github.com/repos/$OWNER/$NAME/releases/tags/$TAG" `
            -Headers $headers -Method Get -ErrorAction Stop
        $releaseId = $release.id
        Write-Host "Release '$TAG' already exists (id=$releaseId)."
    }
    catch {
        if ($_.Exception.Response.StatusCode.value__ -eq 404) {
            Write-Host "Release '$TAG' not found - creating it..."
            $body = @{ tag_name = $TAG; name = $TAG; body = "" } | ConvertTo-Json
            $release = Invoke-RestMethod -Uri "https://api.github.com/repos/$OWNER/$NAME/releases" `
                -Headers $headers -Method Post -Body $body -ContentType "application/json" -ErrorAction Stop
            $releaseId = $release.id
            Write-Host "Created release id=$releaseId."
        }
        else {
            Write-Error "Failed to check release: $_"
            exit 1
        }
    }

    # Delete existing asset with the same name so we can re-upload (clobber)
    $assets = Invoke-RestMethod -Uri "https://api.github.com/repos/$OWNER/$NAME/releases/$releaseId/assets" `
        -Headers $headers -Method Get
    $assetName = Split-Path $tarball -Leaf
    $existing = $assets | Where-Object { $_.name -eq $assetName }
    if ($existing) {
        Write-Host "Deleting existing asset '$assetName'..."
        Invoke-RestMethod -Uri "https://api.github.com/repos/$OWNER/$NAME/releases/assets/$($existing.id)" `
            -Headers $headers -Method Delete | Out-Null
    }

    # Upload the tarball
    $uploadUrl = "https://uploads.github.com/repos/$OWNER/$NAME/releases/$releaseId/assets?name=$assetName"
    $fileBytes = [System.IO.File]::ReadAllBytes((Resolve-Path $tarball).Path)
    Write-Host "Uploading $assetName ($([math]::Round($fileBytes.Length/1MB, 2)) MB)..."
    Invoke-RestMethod -Uri $uploadUrl -Headers $headers -Method Post `
        -Body $fileBytes -ContentType "application/octet-stream" -ErrorAction Stop | Out-Null

    Write-Host "Deploy complete: $assetName uploaded to release '$TAG'."
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
        Write-Output "  deploy          Package and upload as a GitHub release (requires GITHUB_TOKEN env var)"
    }
}
