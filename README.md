bp
==

Simple buildpack runner in Go.

Usage
=====

CLI

```
bp -app 'example-app' -buildpack 'http://github.com/heroku/heroku-buildpack-ruby' -src '/user/home/exampleapp'
```

Code
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
