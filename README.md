bp
==

Simple buildpack runner in Go.

Usage
=====

```
bp, err := NewBuildpack(*bpDir)

if err != nil {
  fmt.Println(err)
  os.Exit(1)
}

if err := bp.Run(*srcDir); err != nil {
  fmt.Println(err)
  os.Exit(1)
}
```
