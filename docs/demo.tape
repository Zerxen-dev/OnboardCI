# VHS documentation tape for OnboardCI terminal demo
# Render with: vhs docs/demo.tape

Output assets/demo.gif
Output assets/demo.mp4

Set Shell "bash"
Set FontSize 16
Set Width 900
Set Height 550
Set Padding 20
Set Theme "Catppuccin Mocha"

Hide
Type "clear"
Enter
Show

Type "onboardci check --dry-run"
Sleep 500ms
Enter
Sleep 2s

Type "onboardci check --execute --cwd testdata/fixtures/broken-env"
Sleep 500ms
Enter
Sleep 3s
