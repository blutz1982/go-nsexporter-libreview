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

# software disclaimer

This project is subject to this disclaimer:

I do not guarantee that this software is free from defects. This software is provided "as is" and you use
software at your own risk.
I make no warranties regarding performance, merchantability, fitness for a particular purpose, or any
other warranties, express or implied.
Under no circumstances shall I be liable for any direct, indirect, special,
incidental or consequential damages resulting from the use, misuse or inability to use this software,
even if I was warned about the possibility of such damage.