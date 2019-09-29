
Write-Host "usage: initproject.ps1 projectName"
Write-Host "----------------------------------"

# $curDir = (pwd) -split '\n'
# echo $curDir

# test is under the root of project folder
# if ((Test-Path ($curDir + "\initproject.ps1")) -eq $false) {
if ((Test-Path "initproject.ps1") -eq $false) {
    Write-Host "please use this script under the root folder of project, and named with 'initproject.ps1'"
    pause
    return
}
$parent = (((pwd) -split '\n') -split '\\')[-1]
$projectName = $parent
if ($args.Count -lt 1) {
    Write-Host ("projectName not specified, use parent folderName(" + $parent + ") instead")
}
else {
    $projectName = $args[0]
}
Write-Host ("projectName: " + $projectName)
pause

if ($env:GOPATH -eq $null) {
    Write-Host "please set GOPATH"
    pause
    return
}
$goSrc = ($env:GOPATH + "\src")

# $katarasDir = ($goSrc + "\gopkg.in\kataras")
# if ((Test-Path $katarasDir) -eq $false) {
#     Invoke-WebRequest -uri "http://10.104.105.71/app/release_pack/iris.v10.tar.gz" -OutFile ($katarasDir + "/iris.v10.tar.gz")
#     $PopUpWin = new-object -comobject wscript.shell
#     # Write-Host "### please manually decompress iris.v10.tar.gz in current folder ###"
#     # Sleep -mill 1000
#     $PopUpWin.popup("please manually decompress iris.v10.tar.gz in current folder")
#     start $katarasDir
#     pause
# }

go mod init ("gitlab.10101111.com/oped/" + $projectName)

Write-Host "renaming..."
dir *.* -Recurse | foreach {
    $fileName = $_.FullName
    if ($fileName.EndsWith("initproject.ps1") -or $fileName.EndsWith("initproject.sh")) {
        Write-Host ("skiped ->" + $fileName)
    }
    else {
        if ((Select-String "redis" $fileName) -ne $null) {
            # do replace
            Write-Host ("processing -> " + $fileName)
            # Get-Content $fileName | %{Write-Host $_.Replace("redis", $projectName)}
            (Get-Content $fileName) -replace "redis",$projectName | Set-Content $fileName
        }
        if ((Select-String "gopkg.in/kataras/iris.v10" $fileName) -ne $null) {
            (Get-Content $fileName) -replace "gopkg.in/kataras/iris.v10", "github.com/kataras/iris@v10.7.1" | Set-Content $fileName
        }
    }
}

cd cmd
ren "redis" $projectName
cd ..
Remove-Item .git -recurse
Write-Host "all done."
pause
