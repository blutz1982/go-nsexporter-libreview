# go-nsexporter-libreview
Transfer your diabetes data from Nightscout to LibreView.

Based on https://github.com/creepymonster/nightscout-to-libreview nodejs project

# ATTENTION: project under development



## Build
see Makefile


## Usage

```
export data to libreview

Usage:
  nsexport libreview [flags]

Flags:
      --date-from string   start of sampling period
      --date-to string     end of sampling period
  -h, --help               help for libreview
      --min-interval int   filter: min sample interval (minutes) (default 12)

Global Flags:
  -c, --config string   path to config (default "config.yaml")
  -d, --debug           toggle debug
```

