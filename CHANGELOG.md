## [1.5.1] (2024-09-20)

### Feat

* add api-secret support for auth

nigthscout YAML section:
```yaml
nightscout:
  apiToken: XXXXXXX
  apiSecret: XXXXXX
  url: http://your-url
```

If both the apiToken and apiSecret fields are specified, the API-secret value takes precedence

## [1.4.3] (2024-07-02)

### Fix

* corrected parser for glucose entry date type (int64 -> float64)

solves the problem:
```bash
2024-07-02T14:21:13+03:00 FTL An error has accured error="nightscout client: cant retreive list glucose entries: strconv.ParseInt: parsing \"1719918959509.336\": invalid syntax"
```


## [1.4.1] (2024-05-06)

* The field used to sample glucose has been changed.

```
was: dateString - string
stato: date - unixMilliseconds
```

## [1.2.1-beta] (2023-10-17)

* Added experimental commands for creating and viewing some Nightscout objects (hidden commands).
* Maken more informative output of some errors.

## [1.2.0-beta] (2023-09-30)

### New

* Added the ability to export long-acting insulin Lantus and Toujeo from xDrip+ (xDrip+ must be configured for multiple insulin profiles. E.g. very-fast Fiasp insulin and long-acting Lantus insulin)

## [1.1.0-beta] (2023-09-26)

### Fix

* corrected type for carbohydrates (int -> float64)

### New

* Added commands for generating, editing and viewing the configuration file
