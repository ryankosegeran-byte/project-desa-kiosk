Add-Type -AssemblyName System.IO.Compression.FileSystem
$zipPath = "d:\PROJECT\MYPROJECT\project-desa-kiosk\template_surat\kalawat\surat_keterangan_usaha\sku.docx"
$zip = [System.IO.Compression.ZipFile]::OpenRead($zipPath)
$entry = $zip.Entries | Where-Object { $_.FullName -eq "word/document.xml" }
$stream = $entry.Open()
$reader = New-Object System.IO.StreamReader($stream)
$content = $reader.ReadToEnd()
$reader.Close()
$zip.Dispose()

# Extract text content (simplified)
$xml = [xml]$content
$ns = New-Object System.Xml.XmlNamespaceManager($xml.NameTable)
$ns.AddNamespace("w", "http://schemas.openxmlformats.org/wordprocessingml/2006/main")
$paragraphs = $xml.SelectNodes("//w:p", $ns)
foreach ($p in $paragraphs) {
    $texts = $p.SelectNodes(".//w:t", $ns)
    $line = ""
    foreach ($t in $texts) {
        $line += $t.InnerText
    }
    if ($line.Trim()) {
        Write-Output $line
    }
}
