; Inno Setup script for StackMatch CLI
; This script creates a Windows installer that also adds the tool to the system PATH.

[Setup]
AppName=StackMatch CLI
; The AppVersion will be passed in from the GitHub Actions workflow
AppVersion=0.0.0
DefaultDirName={autopf}\StackMatch
PrivilegesRequired=admin ; Required to modify system PATH
OutputBaseFilename=stackmatch-cli-setup
Compression=lzma2
SolidCompression=yes
WizardStyle=modern

[Files]
; The source path will be the location of the built exe in the GitHub workflow
Source: "stackmatch-cli.exe"; DestDir: "{app}"; Flags: ignoreversion

[Icons]
Name: "{commonprograms}\StackMatch CLI"; Filename: "{app}\stackmatch-cli.exe"
Name: "{commonprograms}\Uninstall StackMatch CLI"; Filename: "{uninstallexe}"

[Registry]
; Add the application's directory to the system PATH environment variable.
Root: HKLM; Subkey: "SYSTEM\CurrentControlSet\Control\Session Manager\Environment"; ValueType: expandsz; ValueName: "Path"; ValueData: "{olddata};{app}"; Check: NeedsAddPath('{app}')

[Code]
// Helper function to prevent adding the path if it already exists.
function NeedsAddPath(Param: string): boolean;
var
  OldPath: string;
begin
  if not RegQueryStringValue(HKEY_LOCAL_MACHINE, 'SYSTEM\CurrentControlSet\Control\Session Manager\Environment', 'Path', OldPath) then
  begin
    Result := True;
    exit;
  end;
  // Check if the path is already present
  Result := Pos(Param, OldPath) = 0;
end;
