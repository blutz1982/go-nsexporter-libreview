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
      --date-from string       Start of sampling period
      --date-offset string     Start of sampling period with current time offset. Set in duration (e.g. 24h or 72h30m). Ignore --date-from and --date-to flags
      --date-to string         End of sampling period
      --dry-run                Do not post measurement to LibreView
  -h, --help                   help for libreview
      --last-ts-file string    Path to last timestamp file (for example ./last.ts )
      --measurements strings   measurements to upload (default [scheduledContinuousGlucose,unscheduledContinuousGlucose,insulin])
      --min-interval string    Filter: minimum sample interval (duration) (default "10m10s")
      --scan-frequency int     Average scan frequency (minutes). e.g. scan internal min=avg-30%, max=avg+30% (default 90)
      --set-device             Set this app as main user device. Necessary if the main device was set by another application (e.g. Librelink) (default true)
      --ts-layout string       Timestamp layout for --date-from and --date-to flags. More https://go.dev/src/time/format.go (default "2006-01-02")

Global Flags:
  -c, --config string     path to config (default "config.yaml")
  -d, --debug             toggle debug
      --timezone string   override timezone

```

# Explanation of application flags

flags **--date-from** and **--date-to** set the sampling range from Nightscout. If flag **--date-from** is not specified, then its value will be the beginning of the current day. If flag **--date-to** is not specified, then its value will be the current time. The timestamp layout for these flags determined by  **--ts-layout** flag.

flag **--date-offset** determines the time offset (backward) relative to the current time. In other words, the start of the sample will be the current time minus the specified offset, the end of the sample will be the current time.


flag **--last-ts-file** determines the path to the file in which the time stamp of the last exported glucose entry will be stored (or is already stored). When used, all subsequent export operations will exclude all records with a date preceding this timestamp.

flag **--measurements** determines a set of metrics that should be exported to LibreView.


# config

see config.yaml

# example

```bash

# time range with set specify timestamp layout
nsexport libreview --config config.yaml --date-from='2023-09-08T10:50' --date-to='2023-09-08T18:00' --ts-layout="2006-01-02T15:04"

#  time offset with last timestamp file
nsexport libreview --config config.yaml --date-offset=24h --last-ts-file=./last.ts

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