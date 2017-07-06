生成事件->后期生成事件：
$(SolutionDir)copylist.bat $(SolutionDir)/bin/Win32/$(Configuration)

copylist.bat:
cd projdir
xcopy config "%1" /s /y ## 复制config目录下的所有文件和目录到copylist.bat的第一个参数("%1")指定的目录

cd lib/win
for /R %%v in (*.dll) do copy "%%v" "%1"
