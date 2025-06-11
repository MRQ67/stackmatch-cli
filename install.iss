; StackMatch CLI Inno Setup Script

[Setup]
AppName=StackMatch CLI
AppVersion=0.0.0 ; This will be replaced by the workflow
DefaultDirName={autopf}\StackMatch CLI
DefaultGroupName=StackMatch CLI
UninstallDisplayIcon={app}\stackmatch-cli.exe
WizardStyle=modern
OutputBaseFilename=stackmatch-cli-setup
Compression=lzma2
SolidCompression=yes

[Languages]
Name: "english"; MessagesFile: "compiler:Default.isl"

[Files]
Source: "stackmatch-cli.exe"; DestDir: "{app}"; Flags: ignoreversion

[Icons]
Name: "{group}\StackMatch CLI"; Filename: "{app}\stackmatch-cli.exe"
Name: "{group}\Uninstall StackMatch CLI"; Filename: "{uninstallexe}"

[Run]
Filename: "{app}\stackmatch-cli.exe"; Description: "Launch StackMatch CLI"; Flags: nowait postinstall skipifsilent

[Code]
const
    ModPathName = 'Path';
    ModPathType = 'system';

procedure AddToPath(Path: string);
var
    Paths: string;
begin
    if not RegQueryStringValue(HKEY_LOCAL_MACHINE, 'System\CurrentControlSet\Control\Session Manager\Environment', ModPathName, Paths) then
    begin
        Paths := '';
    end;

    if Pos(';' + Path, ';' + Paths) = 0 then
    begin
        if Paths <> '' then
        begin
            Paths := Paths + ';';
        end;
        Paths := Paths + Path;
        RegWriteStringValue(HKEY_LOCAL_MACHINE, 'System\CurrentControlSet\Control\Session Manager\Environment', ModPathName, Paths);
    end;
end;

procedure InitializeWizard();
begin
    AddToPath(ExpandConstant('{app}'));
end;
