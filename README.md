# go-nsexporter-libreview
Transfer your diabetes data from Nightscout to LibreView.

Based on https://github.com/creepymonster/nightscout-to-libreview nodejs project

# ATTENTION: project under development



## Build
see Makefile

## Usage

```
./nsexport libreview --help
export data to libreview

Usage:
  nsexport libreview [flags]

Flags:
      --date-from string     start of sampling period
      --date-to string       end of sampling period
      --dry-run              dont post measurement to libreview
  -h, --help                 help for libreview
      --min-interval int     filter: min sample interval (minutes) (default 12)
      --scan-frequency int   average scan frequency (minutes). e.g. scan internal min=avg-30%, max=avg+30% (default 90)

Global Flags:
  -c, --config string   path to config (default "config.yaml")
  -d, --debug           toggle debug
```

