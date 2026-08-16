param(
    [Parameter(Mandatory = $true)] [string] $InputDirectory,
    [Parameter(Mandatory = $true)] [string] $OutputDirectory
)

$ErrorActionPreference = 'Stop'
$inputRoot = (Resolve-Path -LiteralPath $InputDirectory).Path
$outputRoot = [System.IO.Path]::GetFullPath((Join-Path (Get-Location) $OutputDirectory))
[System.IO.Directory]::CreateDirectory($outputRoot) | Out-Null
$pdfRoot = Join-Path $outputRoot '_pdfs'
[System.IO.Directory]::CreateDirectory($pdfRoot) | Out-Null

$word = New-Object -ComObject Word.Application
$word.Visible = $false
$word.DisplayAlerts = 0
$word.AutomationSecurity = 3

try {
    $index = 0
    Get-ChildItem -LiteralPath $inputRoot -Filter '*.docx' | ForEach-Object {
        $index += 1
        $safeName = $_.BaseName -replace '[^A-Za-z0-9._-]', '_'
        $target = Join-Path $outputRoot $safeName
        [System.IO.Directory]::CreateDirectory($target) | Out-Null
        $pdf = Join-Path $pdfRoot ("doc-{0:D2}.pdf" -f $index)
        Write-Output "Rendering $($_.Name)"
        $document = $word.Documents.Open($_.FullName, $false, $true, $false)
        try {
            $document.SaveAs2($pdf, 17)
        }
        finally {
            $document.Close($false)
        }
        & pdftoppm -png -r 130 $pdf (Join-Path $target 'page')
        if ($LASTEXITCODE -ne 0) {
            throw "Rasterization failed for $($_.FullName)"
        }
        Write-Output "$($_.Name) -> $target"
    }
}
finally {
    $word.Quit()
    [System.Runtime.InteropServices.Marshal]::ReleaseComObject($word) | Out-Null
}
