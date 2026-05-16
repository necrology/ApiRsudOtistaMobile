param(
  [string] $MysqlDumpPath = "C:\xampp\mysql\bin\mysqldump.exe",
  [string] $Database = "apirusdotistamobile",
  [string] $User = "root",
  [string] $Password = "",
  [string] $OutputPath = "docker\mysql\init\001_apirusdotistamobile.sql"
)

$ErrorActionPreference = "Stop"

if (-not (Test-Path -LiteralPath $MysqlDumpPath)) {
  throw "mysqldump tidak ditemukan di $MysqlDumpPath"
}

$outputDirectory = Split-Path -Parent $OutputPath
if (-not (Test-Path -LiteralPath $outputDirectory)) {
  New-Item -ItemType Directory -Path $outputDirectory | Out-Null
}

$args = @(
  "--single-transaction",
  "--routines",
  "--triggers",
  "--events",
  "--default-character-set=utf8mb4",
  "-u", $User
)

if ($Password -ne "") {
  $args += "-p$Password"
}

$args += "--databases"
$args += $Database

& $MysqlDumpPath @args | Set-Content -Path $OutputPath -Encoding utf8

Write-Host "Dump database dibuat: $OutputPath"
