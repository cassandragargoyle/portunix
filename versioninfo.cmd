REM inspiration gptchat and https://learn.microsoft.com/en-us/windows/win32/menurc/versioninfo-resource#examples
# Create a resource file (portunix.rc) with version information
echo 1 VERSIONINFO > portunix.rc
echo FILEVERSION 1,0,0,0 >> portunix.rc
echo PRODUCTVERSION 1,0,0,0 >> portunix.rc
echo FILEFLAGSMASK 0x3fL >> portunix.rc
echo FILEFLAGS 0x0L >> portunix.rc
echo FILEOS 0x4L >> portunix.rc
echo FILETYPE 0x1L >> portunix.rc
echo FILESUBTYPE 0x0L >> portunix.rc
echo BEGIN >> portunix.rc
echo   BLOCK "StringFileInfo" >> portunix.rc
echo   BEGIN >> portunix.rc
echo     BLOCK "040904e4" >> portunix.rc
echo     BEGIN >> portunix.rc
echo       VALUE "CompanyName", "Portunix" >> portunix.rc
echo       VALUE "FileDescription", "Portunix install and config applications" >> portunix.rc
echo       VALUE "FileVersion", "1.0.1" >> portunix.rc
echo       VALUE "InternalName",     "shconfig" >> portunix.rc
echo       VALUE "LegalCopyright",   "Portunix"  >> portunix.rc
echo       VALUE "LegalTrademarks1", "" >> portunix.rc
echo       VALUE "LegalTrademarks2", "" >> portunix.rc
echo       VALUE "OriginalFilename", "portunix.exe" >> portunix.rc
echo       VALUE "ProductName",      "Portunix" >> portunix.rc
echo       VALUE "ProductVersion", "1.0.1" >> portunix.rc
echo     END >> portunix.rc
echo   END >> portunix.rc
echo   BLOCK "VarFileInfo" >> portunix.rc
echo   BEGIN >> portunix.rc
echo     VALUE "Translation", 0x409, 1200 >> portunix.rc
echo   END >> portunix.rc
echo END >> portunix.rc
echo 2 ICON "home.ico" >> portunix.rc

SET PATH=D:\msys64\opt\bin\;%PATH%
SET PATH=D:\msys64\usr\bin\;%PATH%
SET PATH=D:\msys64\mingw64\bin\;%PATH%

REM i686-w64-mingw32-windres.exe -i portunix.rc -O coff -o portunix.syso
REM x86_64-w64-mingw32-windres.exe -i portunix.rc -O coff -o shconfig.syso
windres.exe -i portunix.rc -O coff -o portunix.syso

rsrc -manifest portunix.exe.manifest -o portunix.exe
